package gapi

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/IfanTsai/go-lib/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/pb"
	"github.com/ifantsai/simple-bank-api/util"
	"github.com/pkg/errors"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/protobuf/encoding/protojson"
)

// GatewayServer serves HTTP requests for our banking service.
type GatewayServer struct {
	*GRPCServer
	server *http.Server
}

// NewGatewayServer creates a new gateway server and setup routing.
func NewGatewayServer(config util.Config, store db.Store, address string) (*GatewayServer, error) {
	grpcServer, err := NewGRPCServer(config, store, address)
	if err != nil {
		return nil, errors.Wrap(err, "cannot new grpc server")
	}

	return &GatewayServer{
		GRPCServer: grpcServer,
	}, nil
}

// Start runs the gateway server on a specific address.
func (s *GatewayServer) Start() error {
	jsonLogger := logger.NewJSONLogger(
		logger.WithFileRotationP("./logs/simple-bank.log"),
		logger.WithEnableConsole(),
	)

	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, s.GRPCServer); err != nil {
		return errors.Wrap(err, "cannot register grpc handler")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	fs := http.FileServer(http.Dir("./doc/"))
	mux.Handle("/doc/", http.StripPrefix("/doc/", fs))
	mux.Handle("/swagger/", httpSwagger.Handler(httpSwagger.URL("/doc/swagger/simple_bank.swagger.json")))

	server := &http.Server{
		Addr:              s.address,
		Handler:           HTTPLogger(jsonLogger)(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	s.server = server

	log.Println("gateway server is listening on", s.address)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "failed to start gateway server")
	}

	return nil
}

// Stop stops the gateway server.
func (s *GatewayServer) Stop(ctx context.Context) error {
	return errors.Wrap(s.server.Shutdown(ctx), "failed to shutdown gateway server")
}
