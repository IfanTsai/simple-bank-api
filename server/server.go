package server

import (
	"context"
	"log"
	"time"

	"github.com/IfanTsai/go-lib/process"
	"github.com/pkg/errors"
)

type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

func Run(servers ...Server) {
	for _, server := range servers {
		go func(server Server) {
			if err := server.Start(); err != nil {
				log.Fatal("cannot start server:", err)
			}
		}(server)
	}

	if err := process.GracefulShutdown(
		func(ctx context.Context) error {
			var errRet error
			for _, server := range servers {
				if err := server.Stop(ctx); err != nil {
					if errRet == nil {
						errRet = err
					} else {
						errRet = errors.WithMessagef(errRet, err.Error())
					}
				}
			}

			return errRet
		}, time.Second*10); err != nil {
		log.Fatal("forced to stop server")
	}
}
