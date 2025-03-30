package gRPC

import (
	"context"
	"fmt"
	"google.golang.org/grpc/status"
	pb "metrics/internal/server/proto"
)

type GRPCClient struct {
	client pb.HandlersClient
}

type GRPCRequest struct {
	request *pb.PostUpdatesRequest
}

func NewClient(client pb.HandlersClient) *GRPCClient {
	return &GRPCClient{
		client: client,
	}
}

func (g GRPCClient) PostUpdates(ctx context.Context, requestData []byte) error {
	resp, err := g.client.PostUpdates(ctx, &pb.PostUpdatesRequest{Metric: requestData})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			fmt.Println(e.Code(), e.Message())
		} else {
			fmt.Printf("Can't parse error: %s\n", err.Error())
		}
	}

	return fmt.Errorf("PostUpdates: %w", resp.GetError())
}
