package transportgrpc

import (
	"net"

	"github.com/IbadT/business_bank_back/pkg/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func RunGRPCServer(handler *Handler, port string) (*grpc.Server, error) {
	if port == "" {
		port = "9090"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logrus.WithError(err).Error("Failed to listen")
		return nil, err
	}

	grpcServer := grpc.NewServer()

	// Регистрируем все сервисы
	proto.RegisterGenerateServiceServer(grpcServer, handler)
	proto.RegisterUserServiceServer(grpcServer, handler)
	proto.RegisterTransactionServiceServer(grpcServer, handler)
	proto.RegisterHolidayServiceServer(grpcServer, handler)
	proto.RegisterGatewayServiceServer(grpcServer, handler)
	proto.RegisterBaseAmountServiceServer(grpcServer, handler)
	proto.RegisterBreakdownServiceServer(grpcServer, handler)
	proto.RegisterBalanceServiceServer(grpcServer, handler)

	logrus.Infof("✓ gRPC server started on port %s", port)

	// Запускаем сервер в отдельной goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logrus.WithError(err).Error("gRPC server error")
		}
	}()

	return grpcServer, nil
}
