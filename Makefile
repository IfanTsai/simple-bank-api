DB_URL="postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"

postgres:
	docker run --name postgres12 -p 5432:5432 \
		-e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=simple_bank -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path db/migration -database $(DB_URL) -verbose up

migrateup1:
	migrate -path db/migration -database $(DB_URL) -verbose up 1

migratedown:
	migrate -path db/migration -database $(DB_URL) -verbose down

migratedown1:
	migrate -path db/migration -database $(DB_URL) -verbose down 1

db_docs:
	dbdocs build doc/db.dbml

db_scheme:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover ./...
	rm -rf api/logs

mock:
	mockgen -package mockdb -destination db/mock/store.go  github.com/ifantsai/simple-bank-api/db/sqlc Store

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
    --openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
    proto/*.proto

server:
	go run main.go

evans:
	evans --host localhost --port 9090 -r repl
deps:
	go get github.com/kyleconroy/sqlc/cmd/sqlc
	curl -L https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -
	echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ $$(lsb_release -sc) main" > /etc/apt/sources.list.d/migrate.list
	apt update && apt install -y migrate

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 deps sqlc test server mock dbdocs dbscheme proto evans
