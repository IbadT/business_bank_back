package transportgrpc

import (
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func RunGRPCServer() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logrus.WithError(err).Error("Failed to listen")
		return err
	}

	grpcServer := grpc.NewServer()

	logrus.Info("GRPC server started on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		logrus.WithError(err).Error("Failed to serve")
		return err
	}

	return nil
}
