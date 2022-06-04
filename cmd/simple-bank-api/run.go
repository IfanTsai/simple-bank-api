package main

import (
	"database/sql"
	"log"

	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/gapi"
	"github.com/ifantsai/simple-bank-api/server"
	"github.com/ifantsai/simple-bank-api/util"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load configurations:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)

	grpcServer, err := gapi.NewGRPCServer(config, store, config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot new gRPC server:", err)
	}

	gatewayServer, err := gapi.NewGatewayServer(config, store, config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot new gateway server:", err)
	}

	server.Run(grpcServer, gatewayServer)
}
