# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2026-01-08
### Added
- **Infrastructure:** Added `Dockerfile` using a multi-stage build (Alpine-based) to containerize the API, enabling consistent environments for E2E and load tests.

### Changed
- **Concurrency Model:** Refactored `PaymentsRepository` to use **Go Channels** and the **Monitor Pattern** instead of `sync.Mutex`. State is now owned by a single monitor goroutine, eliminating race conditions by serializing read/write access via `select`.
- **Resilience:** Implemented **Graceful Shutdown** using `golang.org/x/sync/errgroup` and `context.WithCancel`. The server now waits for active connections to finish before closing port 8090 on `SIGTERM`.

### Fixed
- **Testing:** Fixed the load test suite by attaching the k6 container to the API container's network namespace, ensuring reliable connectivity during execution.

## [1.1.0] - 2026-01-08
**Focus:** Quality Assurance & Testing Infrastructure.

### Added
- **E2E Testing:** Created a shell-based integration suite (`tests/e2e/run.sh`) that orchestrates the Docker composition. It verifies the full lifecycle:
    - Authorization (Odd ending card).
    - Decline (Even ending card).
    - Validation Rejection (Business rules).
    - Upstream Failure (Bank 503 simulation).
- **Load Testing:** Integrated `k6` for performance testing, configured to assert P95 latency < 500ms and error rates < 1%.
- **Makefile:** Added `test-e2e` and `test-load` targets to abstract Docker orchestration commands.

### Changed
- **Validation:** Enforced strict currency validation in `validator.go`, allowing only USD, EUR, and BRL (per the 3-currency limit requirement).

## [1.0.0] - 2026-01-07
**Focus:** Core Payment Flow & Upstream Integration.

### Added
- **Business Logic:** Implemented `PostHandler` to orchestrate the payment flow: Request Validation -> Bank Authorization -> Response persistence.
- **Dependency Injection:** Refactored `api.go` to inject the `BankClient` interface into `PaymentsHandler`, enabling mock-based integration testing without the physical bank simulator.
- **Repository:** Implemented `AddPayment` and `GetPayment` using in-memory slice storage (pre-concurrency refactor).

### Changed
- **API Contract:** The `POST /payments` endpoint now proxies the real Acquiring Bank status (`Authorized`/`Declined`) and generates a UUID for the payment ID.

## [0.4.0] - 2026-01-07
**Focus:** Acquiring Bank Integration.

### Added
- **Bank Client:** Developed a dedicated HTTP client (`internal/bank`) with a 5-second timeout to communicate with the Simulator.
- **Data Mapping:** Implemented DTOs to map domain models (Integers for Date) to Bank API format (`MM/YYYY` strings).

## [0.3.0] - 2026-01-06
### Added
- **Validation Strategy:** Implemented a unified `Validate()` method in `internal/payments/validator.go` enforcing:
    - Luhn-ready Card Number format (numeric, 14-19 chars).
    - Future expiry date checks.
    - Positive amount validation.

### Changed
- **Error Responses:** Standardized validation failures to return `400 Bad Request` with specific error messages, distinguishing them from `402 Payment Required` or `502 Bad Gateway`.

## [0.2.0] - 2026-01-06
**Focus:** Security & PCI Compliance concepts.

### Added
- **Security Decisions:**
    - **Transient PAN:** The full Primary Account Number (PAN) is accepted in the request but **never persisted**.
    - **Tokenization/Masking:** Only the last 4 digits (`CardNumberLastFour`) are stored in the repository.
    - **Data Types:** Converted `cvv` and `card_number` to strings to preserve leading zeros and prevent integer overflows.

## [0.1.0] - 2026-01-06
**Focus:** Architecture.

### Changed
- **Project Structure:** Adopted a **Package-by-Feature** architecture (`internal/payments`, `internal/bank`) rather than Layer-based (`handlers`, `models`), improving cohesion and modularity.