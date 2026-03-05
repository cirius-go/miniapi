# Implementation Tasks: Group & Route Package Testing Updates

**Feature Branch**: `001-group-route-tests`
**Created**: March 5, 2026

## Strategy

The implementation will focus on repairing the currently broken `route` package test suite to adapt to the new `RouteBuilder` API, followed by establishing a comprehensive BDD-style test suite for the `group` package to ensure 100% coverage of route grouping logic, path concatenation, and configuration propagation.

### Dependencies & Order

- **Phase 1 (Setup)**: Ensure all dependencies are available and generate any necessary new mocks.
- **Phase 2 (Foundational)**: N/A - Mocks and basic setups are largely handled by the existing `Taskfile.yaml` and `testing_utils_test.go` from the previous route testing phase.
- **Phase 3+ (User Stories)**: 
    - Execute User Story 1 (Route Test Updates) first to stabilize the build.
    - Execute User Story 2 (Group Package Tests) after the route package is stable.

### Parallel Execution Examples

- Tasks within **Phase 4 (User Story 2)** can be executed in parallel (e.g., writing tests for path concatenation vs. security propagation) as they target different methods within the `group` package.

## Phase 1: Setup

Goal: Ensure testing infrastructure is ready, specifically generating any new mocks required for `group` testing.

- [x] T001 Update `Taskfile.yaml` to include `miniapi.Route` and `miniapi.Group` in the mock generation command if they don't already exist.
- [x] T002 Run `task mock` to generate the new interface mocks in the `mocks/` directory.

## Phase 3: User Story 1 - Reliable Route Test Verification

**Goal**: Fix the broken `route` test suite to reflect recent API changes (like `RouteBuilder`) and ensure it runs successfully.
**Independent Test Criteria**: The `route` package tests compile and pass without errors.

- [x] T003 [US1] Update `route/route_test.go` to resolve any undefined method errors resulting from the interface changes (e.g., removing calls to `SetOperation` or `ReqType` if they moved to `RouteBuilder`).
- [x] T004 [US1] Review and update `route/builder_test.go` if necessary to ensure it covers the actual `RouteBuilder` implementation correctly.
- [x] T005 [US1] Run `go test -v ./route/...` and verify all tests pass with 0 failures.

## Phase 4: User Story 2 - Comprehensive Group Package Verification

**Goal**: Implement a robust automated test suite for the `group` package ensuring grouping, pathing, and configuration propagation work flawlessly.
**Independent Test Criteria**: The `group` package tests execute, verifying deep hierarchies and inherited configurations.

- [x] T006 [P] [US2] Implement GoConvey BDD tests in `group/group_test.go` to verify basic group creation (`NewGroup`, `SetPrefix`) and prefix calculation during `Build`.
- [x] T007 [P] [US2] Implement GoConvey BDD tests in `group/group_test.go` to verify adding routes (`AddRoutes`) and sub-groups (`AddGroups`) and that `Build` correctly compiles all routes.
- [x] T008 [P] [US2] Implement GoConvey BDD tests in `group/group_test.go` to verify that `Binder` and `ErrorEncoder` are properly stored and propagated during `Build`.
- [x] T009 [P] [US2] Implement GoConvey BDD tests in `group/group_test.go` to verify that middlewares (`AddMiddlewares`, `Middlewares`, `SetMiddlewares`) are properly stored and managed on the group.
- [x] T010 [P] [US2] Implement GoConvey BDD tests in `group/group_test.go` to verify that route modifiers (`AddModifiers`, `Modifiers`) are properly stored and managed.
- [x] T011 [P] [US2] Implement GoConvey BDD tests in `group/group_test.go` to verify that shared configurations (middlewares, modifiers) are correctly inherited and applied during `Build`.
- [x] T012 [P] [US2] Implement tests for edge cases in `group/group_test.go`, such as handling empty path segments or adding nil routes/modifiers/middlewares.

## Phase 5: Polish

Goal: Final verification of coverage and ensuring the combined suite runs quickly.

- [x] T013 Run `task test-cover` specifically for the `./group/...` package and assert a minimum of 90% statement coverage (aiming for 100%).
- [x] T014 Run the combined suite `go test -v ./group/... ./route/...` and verify it completes in under 2 seconds.
