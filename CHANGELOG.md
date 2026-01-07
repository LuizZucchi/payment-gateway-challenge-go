# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-07
**Focus:** Core Payment Flow, Testing & Documentation.

### Added
- **Payment Flow Logic:** Fully implemented `PostHandler` to orchestrate validation, bank communication, and persistence.
- **Integration Tests:** Added comprehensive integration tests using mocks to verify the full flow (Success, Decline, Bank Error).
- **Wiring:** Updated dependency injection in `api` package to connect the Bank Client with the Payments Handler.

### Changed
- **API Response:** The `POST /payments` endpoint now returns real status from the bank (`Authorized` or `Declined`) instead of stubbed responses.
- **Documentation:** Updated `README.md` with architectural decisions, testing guide, and execution instructions.

## [0.4.0] - 2026-01-07
**Focus:** Acquiring Bank Integration.

### Added
- **Bank Client:** Implemented HTTP client in `internal/bank` to communicate with the Simulator.
- **DTO Mapping:** Added logic to transform domain models into bank-specific formats (e.g., merging `Month/Year` integers into `MM/YYYY` string).

### Decisions
- **Dependency Injection:** The Bank Client is injected into the Payments Handler, allowing for easy mocking during unit tests without needing the actual simulator running.

## [0.3.0] - 2026-01-06

### Added
- **Validation Logic:** Created `validator.go` in the `payments` package to handle business rules validation (Currency, Amount, Card Number, CVV, Expiry Date).
- **Unit Tests:** Added black-box testing for the validator in `validator_test.go` covering all rejection scenarios.

### Changed
- **Data Model:** Updated `CardNumber` field in `PostPaymentRequest` to `string` type to support full PAN length validation and handling.
- **Handlers:** Implemented validation logic within the `PostHandler`. Invalid requests now return `400 Bad Request`.
- **Tests:** Updated handler tests to mock/simulate validation scenarios and accommodate the new data types.

## [0.2.0] - 2026-01-06
**Focus:** Data Modeling, Security & API Contracts.

### Added
- **Sensitive Data Handling (Security):**
    - **Decision:** The `PostPaymentRequest` accepts the full `card_number`, but the full PAN is transient. It is never persisted nor returned in API responses. Only the last 4 digits are stored.
    - **Rationale:** The Gateway acts as a proxy to the Acquiring Bank (which requires the full PAN), but we strictly adhere to security best practices by minimizing data exposure in our persistence layer.

- **Data Types Strategy:**
    - **Decision:** Converted `cvv` and `card_number` fields to `string` in the domain models.
    - **Rationale:**
        - `cvv`: Preserves leading zeros (e.g., "012") which would be lost as integers.
        - `card_number`: Avoids integer overflow issues and facilitates length/Luhn validation logic.

- **RESTful Error Handling:**
    - **Decision:** `GET /payments/{id}` returns `404 Not Found` when the payment does not exist (instead of `204 No Content`).
    - **Rationale:** Adheres to standard RESTful semantics. While `204` implies success with an empty body, `404` correctly indicates the client requested a resource that could not be located.

## [0.1.0] - 2026-01-06
**Focus:** Architectural Foundation & Project Structure.

### Changed
- **Architectural Pattern (Package by Feature):**
    - **Decision:** Refactored the project structure from technical layers (e.g., `handlers/`, `repository/`, `models/`) to domain-centric packages (e.g., `internal/payments/`).
    - **Rationale:**
        - **High Cohesion:** All logic related to payments (entities, storage, validation, transport) is located in a single place, making the code easier to navigate and understand.
        - **Encapsulation:** Allows better control over what is exported publicly vs. what is internal to the payment domain.
        - **Scalability:** Facilitates future extraction into microservices if the domain grows, as the dependencies are already isolated by feature.
        - **Customization:** Enables flexibility; not every package is "forced" to follow a tight layered approach, allowing complexity to grow only where needed.