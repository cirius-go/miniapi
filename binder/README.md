# Binder

The `Binder` package provides a robust and flexible HTTP request binding mechanism for the MiniAPI framework. It extracts parameters from the HTTP request (path, query, headers, and body) and maps them to a Go struct.

## Overview

The core of the package is the `Binder` type, which implements the `miniapi.Binder` interface. It uses reflection and struct tags to populate fields from incoming requests.

### Key Features
- **Path Parameters:** Extracted via `path` struct tags.
- **Query Parameters:** Extracted via `query` struct tags.
- **Header Parameters:** Extracted via `header` struct tags.
- **Body Binding:** Automatically parses the request body based on Content-Type (JSON supported by default).
- **Content Negotiation:** Supports `MarshalResponse` to send back JSON or plain text based on the `Accept` header.

## Configuration

You can configure the binder using the `Config` struct:

```go
type Config struct {
	MaxBodySize           int64
	DisallowUnknownFields bool
}
```

A default configuration is available via `DefaultConfig()`, which sets `MaxBodySize` to 1MB and allows unknown fields.

## Usage

Create a new binder instance and use it to bind requests or marshal responses:

```go
// Initialize Binder
b := binder.New(binder.DefaultConfig())

// Bind a request
err := b.BindRequest(ctx, &myRequestStruct)

// Marshal a response
err = b.MarshalResponse(ctx, &myResponseStruct)
```

## Errors

The package defines several standard errors for common binding failures:
- `ErrTypeMismatch`: A value cannot be converted to the target type.
- `ErrPayloadTooLarge`: The request body exceeds the maximum allowed size.
- `ErrUnsupportedMediaType`: The Content-Type or Accept header is not supported.
- `ErrBindingFailed`: A generic failure during the binding process.
