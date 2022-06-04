package main

import (
	"database/sql"
	"log"

	"github.com/ifantsai/simple-bank-api/api"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
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
	srv, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot new server:", err)
	}

	server.Run(srv, config.GRPCServerAddress)
}
