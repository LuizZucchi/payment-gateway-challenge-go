# Instructions for candidates

This is the Go version of the Payment Gateway challenge. If you haven't already read the [README.md](https://github.com/cko-recruitment/) in the root of this organisation, please do so now. 

## Template structure
```
main.go - a skeleton Payment Gateway API
imposters/ - contains the bank simulator configuration. Don't change this
docs/docs.go - Generated file by Swaggo
.editorconfig - don't change this. It ensures a consistent set of rules for submissions when reformatting code
docker-compose.yml - configures the bank simulator
.goreleaser.yml - Goreleaser configuration
```

Feel free to change the structure of the solution, use a different test library etc.

### Swagger
This template uses Swaggo to autodocument the API and create a Swagger spec. The Swagger UI is available at http://localhost:8090/swagger/index.html.

## Running the Project

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- jq (required for E2E tests)

### Application Commands

#### Run API
Starts the API server locally on port 8090.
```bash
make run

```

#### Build Binary

Compiles the application to the `bin/` directory.

```bash
make build

```

### Testing Commands

#### Unit Tests

Runs standard Go unit tests with coverage.

```bash
make test

```

#### End-to-End (E2E) Tests

Executes the integration test suite. This script automatically:

1. Starts the Bank Simulator (Docker).
2. Starts the Payment Gateway API.
3. Runs a sequence of curl requests to validate success, decline, validation error, and bank error scenarios.
4. Cleans up resources.

```bash
make test-e2e

```

#### Load Tests

Executes a performance test using k6 via Docker. This script automatically:

1. Starts the environment (API + Bank Simulator).
2. Runs the load test script using the `grafana/k6` Docker image.
3. Cleans up resources.

```bash
make test-load

```

```

```