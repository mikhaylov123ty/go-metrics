package gRPC

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "metrics/internal/server/proto"
)

type GRPCClient struct {
	client   pb.HandlersClient
	key      string
	attempts int
	interval time.Duration
}

type gRPCRequest struct {
	*pb.PostUpdatesRequest
	metadata.MD
}

func New(client pb.HandlersClient, key string, attempts int, interval time.Duration) *GRPCClient {
	return &GRPCClient{
		key:      key,
		client:   client,
		attempts: attempts,
		interval: interval,
	}
}

func (g *GRPCClient) PostUpdates(ctx context.Context, requestData []byte) error {
	request := &gRPCRequest{
		&pb.PostUpdatesRequest{Metrics: requestData},
		metadata.New(map[string]string{}),
	}

	decorators := []requestOptions{
		withSign(g.key),
		withRealIP(),
	}

	for _, decorator := range decorators {
		decorator(request)
	}

	ctx = metadata.NewOutgoingContext(ctx, request.MD)
	if err := g.doWithRetry(ctx, request); err != nil {
		return fmt.Errorf("PostUpdates: %w", err)
	}

	return nil
}

type requestOptions func(*gRPCRequest) *gRPCRequest

// Middleware для запросов с подписью
func withSign(key string) requestOptions {
	return func(req *gRPCRequest) *gRPCRequest {
		if key != "" {
			h := hmac.New(sha256.New, []byte(key))
			h.Write([]byte(fmt.Sprintf("%s", req.GetMetrics())))
			hash := hex.EncodeToString(h.Sum(nil))

			req.MD.Set("HashSHA256", hash)

		}
		return req
	}
}

func withRealIP() requestOptions {
	return func(req *gRPCRequest) *gRPCRequest {
		interfaces, err := net.InterfaceAddrs()
		if err != nil {
			log.Printf("failed to get interface addresses: %s", err.Error())
		}

		for _, v := range interfaces {
			if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					req.MD.Set("X-Real-IP", ipnet.IP.String())

					break
				}
			}
		}
		return req
	}
}

// TODO THERE IS INTERCEPTORS BEFORE AND AFTER
func (g *GRPCClient) doWithRetry(ctx context.Context, request *gRPCRequest) error {
	var err error
	wait := 1 * time.Second

	for range g.attempts {
		response, err := g.client.PostUpdates(ctx, request.PostUpdatesRequest)
		if err == nil {
			return nil
		}
		fmt.Println("RESPONSE", response.String())
		if e, ok := status.FromError(err); ok {
			fmt.Println(e.Code(), e.Message())
		} else {
			fmt.Printf("Can't parse error: %s\n", err.Error())
		}
		log.Printf("Worker: TODO HERE, retrying after error: %s\n", err.Error())
		time.Sleep(wait)
		wait += g.interval
	}

	return err
}
