# Implementation Tasks: Route Package Quality Assurance

**Feature Branch**: `001-route-tests`
**Created**: March 5, 2026

## Strategy

The implementation will focus on achieving 100% test coverage for the `route` package using a BDD approach with `goconvey`. The process is divided into setting up the necessary testing tools, creating foundational mocks, and then systematically implementing test suites for each user story. 

### Dependencies & Order

- **Phase 1 (Setup)**: Ensure all dependencies (`goconvey`, `mockery`) and task runners are available.
- **Phase 2 (Foundational)**: Generate necessary mocks (`miniapi.Context`, `miniapi.Binder`, `miniapi.ErrorEncoder`).
- **Phase 3+ (User Stories)**: Implement test files (`route_test.go`, `handler_func_test.go`, `builder_test.go`, `response_schema_test.go`) addressing each user story sequentially.

### Parallel Execution Examples

- Test creation tasks within **Phase 4 (User Story 2)** and **Phase 5 (User Story 3)** can largely be executed in parallel as they test separate, non-overlapping functions within the `route` package (e.g., `handler_func.go` vs `builder.go` modifiers).

## Phase 1: Setup

Goal: Ensure all tooling for the test suite is configured correctly.

- [x] T001 Ensure `github.com/smartystreets/goconvey` and `github.com/vektra/mockery/v2` are present in `go.mod`
- [x] T002 Verify `Taskfile.yaml` exists in the root directory with commands for running `mockery` and generating coverage reports

## Phase 2: Foundational

Goal: Establish the core testing components (mocks and data fixtures) shared across the route tests.

- [x] T003 Generate or verify `miniapi.Binder` mocks into `mocks/Binder.go` using `mockery`
- [x] T004 Generate or verify `miniapi.Context` mocks into `mocks/Context.go` using `mockery`
- [x] T005 Create foundational test fixtures in a shared testing utility file or directly within `route/handler_func_test.go` (e.g., `MockRequest`, `MockResponse`, `MockError`)

## Phase 3: User Story 1 - Reliable Route Construction and Definition

**Goal**: Verify that route builders accurately instantiate and persist route configurations without data loss.
**Independent Test Criteria**: A route created with a specific path, method, and OpenAPI metadata exactly reflects those inputs when inspected.

- [x] T006 [US1] Implement GoConvey BDD tests for basic Route construction (Path, Method) in `route/route_test.go`
- [x] T007 [US1] Implement GoConvey BDD tests for Request/Response reflection types in `route/builder_test.go`
- [x] T008 [US1] Implement GoConvey BDD tests for OpenAPI metadata assignment (Summary, Description, Tags) in `route/builder_test.go`
- [x] T009 [US1] Implement GoConvey BDD tests to verify `response_schema.go` builder functions (`JSONSchema`, `ProblemSchema`, `NoContentSchema`) return correct `miniapi.Response` instances in `route/response_schema_test.go`

## Phase 4: User Story 2 - Predictable Request Execution and Binding

**Goal**: Verify that the execution handler seamlessly bridges incoming requests to user logic and formats the output.
**Independent Test Criteria**: A mocked HTTP request flows through the binder, executes the mock business logic, and the result is passed back to the response marshaler.

- [x] T010 [P] [US2] Implement GoConvey BDD tests verifying successful request binding and execution flow in `route/handler_func_test.go`
- [x] T011 [P] [US2] Implement GoConvey BDD tests verifying correct error handling when request binding fails (e.g., returning HTTP 400 or 500 based on the default error encoder) in `route/handler_func_test.go`
- [x] T012 [P] [US2] Implement GoConvey BDD tests verifying correct error handling when the user's business logic returns an error in `route/handler_func_test.go`
- [x] T013 [P] [US2] Implement GoConvey BDD tests verifying successful response marshaling and default status code assignment (HTTP 200) in `route/handler_func_test.go`
- [x] T014 [P] [US2] Implement tests for edge cases like nil handler functions or custom ErrorEncoder injection in `route/handler_func_test.go`

## Phase 5: User Story 3 - Extensible Middleware and Modifiers

**Goal**: Verify that routes correctly aggregate and apply middleware and modifiers in the expected order.
**Independent Test Criteria**: Adding multiple distinct middleware functions to a route results in a properly chained execution pipeline.

- [x] T015 [P] [US3] Implement GoConvey BDD tests for adding and retrieving middleware lists from a Builder in `route/builder_test.go`
- [x] T016 [P] [US3] Implement GoConvey BDD tests verifying that the `chain` function correctly executes middleware in the order they were added in `route/builder_test.go`
- [x] T017 [P] [US3] Implement GoConvey BDD tests for adding and resolving route modifiers (like security overrides) without mutating the original builder state in `route/builder_test.go`
- [x] T018 [P] [US3] Implement tests for edge cases like applying nil or empty middleware slices in `route/builder_test.go`

## Phase 6: Polish

Goal: Final verification of coverage and ensuring coding standards are met.

- [x] T019 Run `task test-cover` specifically for the `./route/...` package and assert 100% statement coverage.
- [x] T020 Review the `route` package code to confirm that naming conventions (`JSONSchema`, `ProblemSchema`, `NoContentSchema`) are idiomatic and consistent with FR-007.
