package gapi

import (
	"context"
	"net"

	"github.com/IfanTsai/go-lib/user/token"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/pb"
	"github.com/ifantsai/simple-bank-api/util"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server serves gRPC requests for our banking service.
type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	server     *grpc.Server
}

// NewServer creates a new gRPC server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create token")
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}

// Start runs the HTTP server on a specific address.
func (s *Server) Start(address string) error {
	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, s)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return errors.Wrap(err, "cannot new tcp listener")
	}

	s.server = grpcServer

	return errors.Wrap(grpcServer.Serve(listener), "failed to Gin server")
}

// Stop stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.server.GracefulStop()

	return nil
}
