package main

import (
	"database/sql"
	"log"

	"github.com/IfanTsai/go-lib/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/gapi"
	"github.com/ifantsai/simple-bank-api/server"
	"github.com/ifantsai/simple-bank-api/util"
	"github.com/ifantsai/simple-bank-api/worker"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
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

	runDBMigration(config.DBMigrationURL, config.DBSource)

	store := db.NewStore(conn)

	workerLogger := logger.NewJSONLogger(
		logger.WithFileRotationP("./logs/worker.log"),
		logger.WithEnableConsole(),
	)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt, worker.WithLogger(workerLogger))

	grpcServer, err := gapi.NewGRPCServer(config, store, config.GRPCServerAddress, taskDistributor)
	if err != nil {
		log.Fatal("cannot new gRPC server:", err)
	}

	gatewayServer, err := gapi.NewGatewayServer(config, store, config.HTTPServerAddress, taskDistributor)
	if err != nil {
		log.Fatal("cannot new gateway server:", err)
	}

	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, worker.WithLogger(workerLogger))

	server.Run(taskProcessor, grpcServer, gatewayServer)
}

func runDBMigration(url string, source string) {
	migration, err := migrate.New(url, source)
	if err != nil {
		log.Fatal("cannot new migration:", err)
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("cannot migrate:", err)
	}

	log.Println("migration done")
}
