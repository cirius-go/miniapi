<!--
Sync Impact Report:
Version change: 1.0.0 → 1.1.0
Modified principles: Initialized MiniAPI principles.
Added sections: Code Quality Principles, Testing Standards, User Experience Consistency, Performance Requirements, Backward Compatibility, Simplicity as a Design Rule.
Templates requiring updates:
- .specify/templates/plan-template.md (✅ aligned Constitution Check)
- .specify/templates/spec-template.md (✅ aligned)
- .specify/templates/tasks-template.md (✅ aligned)
Follow-up TODOs: 
- TODO(RATIFICATION_DATE): Needs exact adoption date for version 1.0.0.
-->

# MiniAPI Constitution

## Purpose

MiniAPI aims to be a minimal, type-safe HTTP framework that produces reliable APIs and accurate OpenAPI specifications.
The framework prioritizes correctness, clarity, performance, and a predictable developer experience.

All contributions must follow the principles defined in this constitution.

---

# 1. Code Quality Principles

## 1.1 Clarity over cleverness

Code must prioritize readability and explicit intent.

Avoid:
* hidden side effects
* implicit behaviors
* overly abstract layers
* unnecessary generics

Prefer simple and predictable implementations.
If a feature requires complex logic, it must be isolated and documented.

---

## 1.2 Deterministic behavior

All framework behavior must be deterministic.

Examples:
* OpenAPI generation must not depend on Go map iteration order.
* Middleware order must be explicitly defined.
* Schema registration must be stable across runs.

Non-deterministic behavior is considered a bug.

---

## 1.3 Minimal surface area

MiniAPI intentionally keeps its API small.

The framework must not expand into:
* dependency injection frameworks
* ORMs
* application business logic
* complex plugin systems

The framework provides infrastructure only.

---

## 1.4 Explicit configuration

Automatic behavior must never override explicit developer intent.

Priority order:
1. explicit route configuration
2. framework inference
3. framework defaults

Explicit configuration always wins.

---

# 2. Testing Standards

## 2.1 Mandatory test coverage

All public functionality must have tests.

Minimum expectations:
* unit tests for internal logic
* integration tests for routing behavior
* OpenAPI spec generation tests

Code without tests must not be merged.

---

## 2.2 Behavior-driven validation

Tests must validate **observable behavior**, not internal implementation.

Examples:
Correct:
* verify generated OpenAPI schema
* verify middleware execution order
* verify response serialization

Incorrect:
* testing private variables
* testing internal helper functions without behavior impact

---

## 2.3 Regression safety

Every bug fix must include a regression test.
This ensures previously fixed issues cannot reappear.

---

## 2.4 Deterministic tests

Tests must be reproducible.

Avoid:
* time-based logic
* random ordering
* environment-dependent results

All tests must pass consistently in CI environments.

---

# 3. User Experience Consistency

MiniAPI's primary users are backend developers.
The framework must provide a consistent and predictable developer experience.

---

## 3.1 Consistent handler model

All handlers must follow the same pattern:
```go
func(ctx context.Context, req *RequestType) (*ResponseType, error)
```

Handlers must not interact directly with:
* http.ResponseWriter
* http.Request

The framework is responsible for HTTP transport details.

---

## 3.2 Consistent response semantics

HTTP responses must follow standard semantics.

Examples:
* 200 for successful retrieval
* 201 for resource creation
* 204 for success without body

A 204 response must never include a response body.

---

## 3.3 Consistent error model

Errors returned by handlers must produce a consistent API error structure.

Example:
```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

The framework must ensure error responses are predictable and documented.

---

## 3.4 Accurate API documentation

Generated OpenAPI specifications must reflect actual runtime behavior.

The framework must never:
* generate schemas that cannot be returned
* hide declared responses
* misrepresent security requirements

Documentation correctness is mandatory.

---

# 4. Performance Requirements

MiniAPI must remain lightweight and efficient.
Performance regressions are unacceptable.

---

## 4.1 Zero unnecessary allocations

Critical request paths must avoid unnecessary allocations.
Reflection and schema generation must occur only during startup or spec generation, not per request.

---

## 4.2 Predictable request overhead

Middleware chains and request processing must remain O(n) relative to middleware count.
Hidden complexity must be avoided.

---

## 4.3 Startup-time reflection

Reflection should occur during initialization rather than request handling whenever possible.
Runtime request paths must remain minimal.

---

## 4.4 No hidden background processes

The framework must not start goroutines, schedulers, or background workers without explicit user configuration.

---

# 5. Backward Compatibility

Public APIs must remain stable.

Breaking changes require:
* major version updates
* migration documentation

Minor releases must remain backward compatible.

---

# 6. Simplicity as a Design Rule

MiniAPI must remain simple enough for developers to understand its full behavior by reading the source code.
If a feature significantly increases complexity, it must be reconsidered.
Framework simplicity is a core goal.

---

# Final Principle

MiniAPI must always favor:
correctness → clarity → performance → convenience
in that order.

---

## Governance

Amendments require documentation, approval, and a migration plan if necessary. All PRs/reviews must verify compliance with this constitution.

**Version**: 1.1.0 | **Ratified**: TODO(RATIFICATION_DATE) | **Last Amended**: 2026-03-05