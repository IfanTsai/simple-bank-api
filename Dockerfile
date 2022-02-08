# Build stage
FROM golang:1.17-alpine3.14 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine:3.14
# fix panic: could not load time location: unknown time zone Asia/Shanghai
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /opt/zoneinfo.zip
ENV ZONEINFO /opt/zoneinfo.zip
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate.linux-amd64 ./migrate
COPY db/migration ./migration
COPY app.env .
COPY start_docker.sh .
COPY wait-for.sh .

EXPOSE 8080
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start_docker.sh" ]
