# Payment Gateway - Technical Design Document

**Author:** Luiz Zucchi
**Date:** January 8, 2026
**Version:** 1.1.1

## 1. Architecture Overview

### 1.1 Package-by-Feature Strategy

The project adopts a **Package-by-Feature** structure (e.g., `internal/payments`, `internal/bank`) rather than a traditional Layered architecture (`controllers`, `services`, `repositories`).

* **Reasoning:** This improves **cohesion**. All logic related to payments—validation, storage, and transport—resides in a single directory. This structure reduces cognitive load when navigating the codebase and facilitates future extraction into microservices, as boundaries are defined by domain rather than technical function.

### 1.2 Dependency Injection (DI)

Dependencies are explicitly defined in struct constructors and wired in `internal/api/api.go`.

* **Implementation:** `PaymentsHandler` depends on the `BankGateway` interface rather than the concrete `BankClient`.
* **Benefit:** This allows the use of a `MockBankGateway` during unit tests (`internal/payments/handler_test.go`), enabling testing of the API logic in isolation without requiring the physical Bank Simulator to be running.

---

## 2. Concurrency & State Management

### 2.1 The Monitor Pattern (Actor Model)

To address the "In-Memory Repository" requirement without race conditions, I implemented a **Monitor Pattern** using Go Channels in `internal/payments/repository.go`.

* **Problem:** The default slice `[]models.PostPaymentResponse` is not thread-safe. Concurrent writes from multiple HTTP requests would cause panics or data corruption. We notice this during load tests
* **Solution:** A single "Monitor" goroutine owns the data slice.
* **Writes:** Handlers send data to an `addChan`. The monitor reads from this channel and appends to the slice.
* **Reads:** Handlers send a request struct (containing a response channel) to `getChan`. The monitor processes the search and sends the result back.


* **Why Channels over Mutex?** While a `sync.RWMutex` would be sufficient for this scale, using channels aligns with Go's philosophy ("Do not communicate by sharing memory; share memory by communicating"). It avoids lock contention issues and ensures sequential processing of state changes.

### 2.2 Graceful Shutdown

The application uses `errgroup` and `context` to handle termination signals (`SIGTERM`, `SIGINT`).

* **Behavior:** When a signal is received, the HTTP server stops accepting new connections but allows in-flight requests to complete before killing the process. This prevents dropped transactions during deployments.

---

## 3. Security & Compliance (PCI-DSS)

### 3.1 PAN Handling

Handling credit card numbers (PAN) requires strict adherence to security standards.

* **Decision:** The application accepts the full PAN to forward it to the Acquiring Bank, but **never persists it**.
* **Storage:** Only the `CardNumberLastFour` (last 4 digits) is stored in the `PostPaymentResponse` struct within the repository.
* **Logging:** The application logger is configured to ensure PANs do not leak into logs (implementation note for production config).

### 3.2 Data Types

* **String vs Integer:** `CardNumber` and `CVV` are treated as `string` types.
* **Reasoning:** Credit card numbers are identifiers, not mathematical values. Storing them as integers can cause overflow (standard `int` might not hold 16 digits on 32-bit systems) and loss of leading zeros in CVVs (e.g., "012" becoming "12").



---

## 4. Testing Strategy

The solution employs a Testing Pyramid approach:

1. **Unit Tests:**
* Focus on business logic (`validator_test.go`) and internal banking logic (`client_test.go`).
* The Bank Client tests use `httptest.NewServer` to mock the external bank API, ensuring tests are fast and deterministic.


2. **Integration Tests:**
* `handler_test.go` tests the wiring between the HTTP layer and the storage/bank interfaces using mocks.


3. **End-to-End (E2E) Tests (`tests/e2e/run.sh`):**
* Black-box testing using `curl` and `jq`.
* Spins up the full Docker environment (API + Mountebank Simulator).
* Validates the system contract from the outside, verifying that the API, Repository, and Bank Client work together correctly in a containerized network.


4. **Load Tests (`tests/load/`):**
* Uses **k6** to simulate high-concurrency traffic.
* Validates that the Concurrency Monitor (Section 2.1) performs under load without race conditions or deadlocks.



---

## 5. API Design Choices

### 5.1 RESTful Semantics

* **GET /payments/{id}:** Returns `200 OK` with the payment details or `404 Not Found` if the ID does not exist. I avoided `204 No Content` for missing resources as `404` is more explicit for client errors.
* **POST /payments:** Returns `200 OK` for both Authorized and Declined transactions (as both are successful *processing* events), but returns `502 Bad Gateway` if the upstream bank is unreachable.

### 5.2 Validation Logic

Validation is centralized in a `Validate()` method (`internal/payments/validator.go`).

* **Strict Whitelisting:** Currencies are strictly validated against a hardcoded list (`USD`, `EUR`, `BRL`) to meet business requirements.
* **Sanitization:** Spaces are stripped from Card Numbers before length validation to improve user experience (accepting "1234 5678...").