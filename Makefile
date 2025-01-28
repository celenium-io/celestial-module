-include .env
export $(shell sed 's/=.*//' .env)

generate:
	go generate -v ./pkg/api ./pkg/module ./pkg/storage

lint:
	golangci-lint run

test:
	go test -p 8 -timeout 60s ./...

.PHONY: generate lint test