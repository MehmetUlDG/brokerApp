.PHONY: proto tidy build run

## proto: regenerate Go code from payment.proto (requires protoc + plugins)
proto:
	protoc \
		--go_out=gen/payment \
		--go_opt=paths=source_relative \
		--go-grpc_out=gen/payment \
		--go-grpc_opt=paths=source_relative \
		-I proto \
		proto/payment.proto

## tidy: download and tidy Go dependencies
tidy:
	go mod tidy

## build: compile the binary
build:
	go build -o bin/payment-service ./cmd/api

## run: run the service (requires .env to be set up)
run:
	go run ./cmd/api
