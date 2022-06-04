package server

import (
	"context"
	"log"
	"time"

	"github.com/IfanTsai/go-lib/process"
	"github.com/pkg/errors"
)

type Server interface {
	Start(address string) error
	Stop(ctx context.Context) error
}

func Run(server Server, address string) {
	go func() {
		if err := server.Start(address); err != nil {
			log.Fatal("cannot start server:", err)
		}
	}()

	if err := process.GracefulShutdown(
		func(ctx context.Context) error {
			return errors.WithMessage(server.Stop(ctx), "failed to stop server")
		}, time.Second*10); err != nil {
		log.Fatal("forced to stop server")
	}
}
