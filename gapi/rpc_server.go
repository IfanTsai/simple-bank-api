package gapi

import (
	"context"
	"log"
	"net"

	"github.com/IfanTsai/go-lib/logger"
	"github.com/IfanTsai/go-lib/user/token"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/pb"
	"github.com/ifantsai/simple-bank-api/util"
	"github.com/ifantsai/simple-bank-api/worker"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer serves gRPC requests for our banking service.
type GRPCServer struct {
	pb.UnimplementedSimpleBankServer
	server          *grpc.Server
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	address         string
	taskDistributor worker.TaskDistributor
}

// NewGRPCServer creates a new gRPC server and setup routing.
func NewGRPCServer(
	config util.Config, store db.Store, address string, taskDistributor worker.TaskDistributor,
) (*GRPCServer, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create token")
	}
	server := &GRPCServer{
		config:          config,
		store:           store,
		address:         address,
		taskDistributor: taskDistributor,
		tokenMaker:      tokenMaker,
	}

	return server, nil
}

// Start runs the gRPC server on a specific address.
func (s *GRPCServer) Start() error {
	jsonLogger := logger.NewJSONLogger(
		logger.WithFileRotationP("./logs/simple-bank.log"),
		logger.WithEnableConsole(),
	)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(GRPCLogger(jsonLogger)),
	)
	pb.RegisterSimpleBankServer(grpcServer, s)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return errors.Wrap(err, "cannot new tcp listener")
	}

	s.server = grpcServer
	log.Println("gRPC server is listening on", s.address)

	return errors.Wrap(grpcServer.Serve(listener), "failed to run gRPC server")
}

// Stop stops the gRPC server.
func (s *GRPCServer) Stop(ctx context.Context) error {
	s.server.GracefulStop()

	return nil
}
