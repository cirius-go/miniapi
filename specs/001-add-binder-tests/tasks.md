# Implementation Tasks: Binder Testing Reliability

**Feature Branch**: `001-add-binder-tests`
**Created**: March 5, 2026

## Strategy

The implementation will follow a strict BDD testing approach to achieve 100% code coverage on the existing `binder` package. The strategy relies on generating mocks for `miniapi.Context` using `mockery`, establishing robust data fixtures, and verifying each feature systematically in GoConvey blocks.

### Dependencies & Order

- **Phase 1 (Setup)**: Prepare dependencies and mock generation.
- **Phase 2 (Foundational)**: Generate the actual `miniapi.Context` mocks and set up shared test fixtures.
- **Phase 3+ (User Stories)**: Implement test suites for each user story sequentially, ensuring each block runs independently.

### Parallel Execution Examples

- **Phase 3 (User Story 1)**, **Phase 4 (User Story 2)**, and **Phase 5 (User Story 3)** tests can be written and executed in parallel once Phase 1 and 2 are complete, as they focus on different aspects of the `binder` behavior (parsing, error handling, content negotiation).

## Phase 1: Setup

Goal: Ensure all required tooling for the new test suite is in place.

- [x] T001 Install or verify `github.com/smartystreets/goconvey` and `github.com/vektra/mockery/v2` in `go.mod`
- [x] T002 Add a standard `Taskfile.yaml` or task runner command for running `mockery` and generating coverage reports in the root directory

## Phase 2: Foundational

Goal: Establish the core testing components (mocks and data fixtures) that will be shared across all user story tests.

- [x] T003 Generate `miniapi.Context` mocks into `mocks/miniapi/Context.go` using `mockery`
- [x] T004 Create foundational test fixtures in `binder/binder_test.go` (e.g., `PrimitiveStruct`, `SliceStruct`, `PointerStruct`, `JSONBodyStruct`, `RawBodyStruct`)

## Phase 3: User Story 1 - Developer Trust in Request Parsing

**Goal**: Verify the binder correctly extracts path, query, header, and body data into structured Go types.
**Independent Test Criteria**: Standard valid requests map perfectly to complex structs containing primitives, slices, and pointers.

- [x] T005 [P] [US1] Implement GoConvey BDD tests for mapping path parameters to primitives in `binder/binder_test.go`
- [x] T006 [P] [US1] Implement GoConvey BDD tests for mapping query parameters (single values and slices) in `binder/binder_test.go`
- [x] T007 [P] [US1] Implement GoConvey BDD tests for mapping header parameters (single values and slices) in `binder/binder_test.go`
- [x] T008 [P] [US1] Implement GoConvey BDD tests for mapping valid JSON bodies to `JSONBodyStruct` in `binder/binder_test.go`
- [x] T009 [P] [US1] Implement GoConvey BDD tests for mapping to pointer fields and verifying allocation logic in `binder/binder_test.go`
- [x] T010 [P] [US1] Implement GoConvey BDD tests for mapping request bodies to raw binary streams (`io.Reader` and `[]byte`) in `binder/binder_test.go`
- [x] T011 [P] [US1] Implement GoConvey BDD tests for missing values (e.g., empty string processing) in `binder/binder_test.go`

## Phase 4: User Story 2 - Resilient Error Handling for Invalid Inputs

**Goal**: Verify the framework gracefully rejects invalid or oversized requests without panicking.
**Independent Test Criteria**: Malformed JSON, wrong data types, and oversized payloads return predictable framework errors.

- [x] T012 [P] [US2] Implement GoConvey BDD tests for invalid primitive type conversions (e.g., string to int) in `binder/binder_test.go`
- [x] T013 [P] [US2] Implement GoConvey BDD tests for request payloads exceeding `MaxBodySize` returning payload-too-large errors in `binder/binder_test.go`
- [x] T014 [P] [US2] Implement GoConvey BDD tests for malformed JSON bodies returning standard binding errors in `binder/binder_test.go`
- [x] T015 [P] [US2] Implement GoConvey BDD tests for unsupported `Content-Type` headers in `binder/binder_test.go`
- [x] T016 [P] [US2] Implement GoConvey BDD tests for rejecting unknown fields when `DisallowUnknownFields` is enabled in `binder/binder_test.go`
- [x] T017 [P] [US2] Implement GoConvey BDD tests for handling `req` not being a pointer to a struct in `binder/binder_test.go`

## Phase 5: User Story 3 - Accurate Content Negotiation and Output

**Goal**: Verify responses are serialized correctly based on client preferences.
**Independent Test Criteria**: Text and JSON content types are respected and formatted properly via `MarshalResponse`.

- [x] T018 [P] [US3] Implement GoConvey BDD tests for marshaling structures to JSON when client `Accept` prefers JSON in `binder/binder_test.go`
- [x] T019 [P] [US3] Implement GoConvey BDD tests for formatting responses to plain text when client `Accept` prefers text in `binder/binder_test.go`
- [x] T020 [P] [US3] Implement GoConvey BDD tests for default JSON formatting when no `Accept` header is present in `binder/binder_test.go`
- [x] T021 [P] [US3] Implement GoConvey BDD tests for marshaling a `nil` response cleanly in `binder/binder_test.go`

## Phase 6: Polish

Goal: Final verification of coverage and cleanup.

- [x] T022 Run `go test -coverprofile=coverage.out ./binder/...` and verify 100% statement coverage is achieved across `binder.go` and `errors.go`
- [x] T023 Refactor any duplicated mock setup logic into helper functions within `binder/binder_test.go`
