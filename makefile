.PHONY: run test build clean check

run:
	go run main.go

test:
	go test -v -cover ./...

build:
	go build -o bin/payment-gateway main.go

lint:
	go fmt ./...
	go vet ./...

check: lint test build