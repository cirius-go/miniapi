# miniapi Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-24

## Active Technologies
- Go (latest supported by project) + `github.com/getkin/kin-openapi` (OpenAPI 3.0 generation), `github.com/smartystreets/goconvey/convey` (Testing) (001-openapi-extractor)
- Go (latest supported by project) + None (HTML strings embedded directly in Go code, fetching assets from public CDNs: `cdn.redoc.ly` and `cdn.jsdelivr.net/npm/@scalar/api-reference`) (002-openapi-ui)
- Go (latest supported by project) + None external (003-route-modifiers)
- Go (latest supported by project) + `github.com/getkin/kin-openapi`, `github.com/smartystreets/goconvey/convey` (004-refactor-openapi-ui)
- N/A (HTML templates embedded in Go code, assets from CDNs) (004-refactor-openapi-ui)
- Go (latest supported by project) + Standard Library (`encoding/json`, `reflect`, `context`, `io`) (005-typed-handler-conversion)
- Go 1.25.5 + Standard Library (`reflect`, `encoding/json`, `net/http`, `io`); `github.com/labstack/echo/v4` (for adapter support); `github.com/smartystreets/goconvey/convey` (for testing). (006-typed-handler-binding)
- Go 1.25.5 + Standard Library (`reflect`, `encoding/json`, `net/http`, `io`); `github.com/vektra/mockery/v2` (for automated mock generation); `github.com/smartystreets/goconvey/convey` (for testing). (007-binder-interface-mock)
- Go 1.25.5 + `github.com/getkin/kin-openapi` v0.133.0, `github.com/smartystreets/goconvey/convey` (testing). (008-fix-openapi-duplication)
- Go (latest supported by project) + `github.com/getkin/kin-openapi` (for documenting security schemes) (001-lib-auth)

- Go (latest supported by project) + Minimal standard library dependencies, GoConvey for testing. (001-openapi-extractor)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go (latest supported by project)

## Code Style

Go (latest supported by project): Follow standard conventions

## Recent Changes
- 010-openapi-security-refactor: Added Go (latest supported by project) + `github.com/getkin/kin-openapi`
- 001-openapi-auth: Added Go (latest supported by project) + `github.com/getkin/kin-openapi`
- 001-lib-auth: Added Go (latest supported by project) + `github.com/getkin/kin-openapi` (for documenting security schemes)


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
