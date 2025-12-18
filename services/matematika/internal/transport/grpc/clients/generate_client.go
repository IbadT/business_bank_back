package clients

import (
	"fmt"

	generatepb "github.com/IbadT/business_bank_back/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGenerateClient(addr string) (generatepb.GenerateServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	client := generatepb.NewGenerateServiceClient(conn)
	return client, conn, nil
}
