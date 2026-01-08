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

test-e2e:
	@echo "Running Integration/E2E Tests..."
	@chmod +x tests/e2e/run.sh
	@cd tests/e2e && ./run.sh

test-load:
	@echo "Running Load Tests with k6..."
	@chmod +x tests/load/run.sh
	@cd tests/load && ./run.sh
