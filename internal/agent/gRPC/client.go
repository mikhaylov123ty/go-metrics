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

	"metrics/pkg"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "metrics/internal/server/proto"
)

// TODO annotations
type GRPCClient struct {
	client   pb.HandlersClient
	attempts int
	interval time.Duration
}

func New(client pb.HandlersClient, attempts int, interval time.Duration) *GRPCClient {
	return &GRPCClient{
		client:   client,
		attempts: attempts,
		interval: interval,
	}
}

func (g *GRPCClient) PostUpdates(ctx context.Context, requestData []byte) error {
	if err := g.doWithRetry(ctx, &pb.PostUpdatesRequest{Metrics: requestData}); err != nil {
		return fmt.Errorf("PostUpdates: %w", err)
	}

	return nil
}

func NewInterceptors(key string) grpc.DialOption {
	interceptors := []grpc.UnaryClientInterceptor{
		withHash(key),
		withRealIP(),
	}

	return grpc.WithChainUnaryInterceptor(interceptors...)
}

// Middleware для запросов с подписью
func withHash(key string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if key != "" {
			h := hmac.New(sha256.New, []byte(key))
			h.Write([]byte(fmt.Sprintf("%s", req.(*pb.PostUpdatesRequest).Metrics)))
			hash := hex.EncodeToString(h.Sum(nil))

			ctx = metadata.AppendToOutgoingContext(ctx, "HashSHA256", hash)

		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func withRealIP() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		interfaces, err := net.InterfaceAddrs()
		if err != nil {
			log.Printf("failed to get interface addresses: %s", err.Error())
		}

		for _, v := range interfaces {
			if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ctx = metadata.AppendToOutgoingContext(ctx, "X-Real-IP", ipnet.IP.String())
					break
				}
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (g *GRPCClient) doWithRetry(ctx context.Context, request *pb.PostUpdatesRequest) error {
	var err error
	wait := 1 * time.Second

	workerID, ok := ctx.Value(pkg.ContextKey{}).(int)
	if !ok {
		log.Printf("gRPC client: failed to get worker ID")
	}

	for range g.attempts {
		_, err = g.client.PostUpdates(ctx, request)
		if err == nil {
			return nil
		}
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.Unavailable:
				log.Printf("Worker: %d, retrying after error: %s\n", workerID, err.Error())
				time.Sleep(wait)
				wait += g.interval
			default:
				return fmt.Errorf("post updates: Code: %s, Message: %s", e.Code(), e.Message())
			}
		} else {
			log.Printf("Can't parse error: %s\n", err.Error())
		}
	}

	return err
}
