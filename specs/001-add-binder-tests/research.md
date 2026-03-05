# Research: Binder Testing Reliability

## Phase 0: Outline & Research

### Testing Framework and Style

**Decision**: Use standard Go `testing` combined with `github.com/smartystreets/goconvey/convey` specifically leveraging its Behavior-Driven Development (BDD) style format.

**Rationale**: `goconvey` explicitly supports BDD style syntax (`Convey`, `So`) which aligns perfectly with the "Given/When/Then" Acceptance Scenarios defined in the feature spec. This forces tests to be descriptive, acts as living documentation, and helps ensure 100% of defined edge cases are clearly accounted for.

**Alternatives considered**: Pure standard library `testing` (can be verbose and less readable for complex nested scenarios).

### Mocking HTTP Requests and Contexts

**Decision**: Use `github.com/vektra/mockery/v2` to generate automated, type-safe mocks for the `miniapi.Context` interface.

**Rationale**: Achieving 100% test coverage requires simulating various deep failure states (like I/O errors when reading a body, or specific HTTP header combinations). `mockery` allows precise injection of these behaviors and returns into the mocked `Context` methods without needing to spin up actual HTTP servers or construct complex fake objects by hand. This keeps tests blisteringly fast and strictly focused on `binder` logic.

**Alternatives considered**: Hand-writing stubs for `miniapi.Context` (error-prone, hard to maintain for 100% coverage edge cases). Using an actual HTTP framework adapter like `echov4` for tests (violates unit testing principles, slow, hard to simulate low-level I/O failures).