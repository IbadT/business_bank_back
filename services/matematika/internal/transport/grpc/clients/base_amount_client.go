package clients

import (
	"fmt"

	baseamountpb "github.com/IbadT/business_bank_back/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewBaseAmountClient(addr string) (baseamountpb.BaseAmountServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	client := baseamountpb.NewBaseAmountServiceClient(conn)
	return client, conn, nil
}
