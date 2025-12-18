package clients

import (
	"fmt"

	balancepb "github.com/IbadT/business_bank_back/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewBalanceClient(addr string) (balancepb.BalanceServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	client := balancepb.NewBalanceServiceClient(conn)
	return client, conn, nil
}
