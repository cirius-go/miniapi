package binder

import "errors"

var (
	// ErrTypeMismatch indicates that a value cannot be converted to the target type.
	ErrTypeMismatch = errors.New("type mismatch")

	// ErrPayloadTooLarge indicates that the request body exceeds the maximum allowed size.
	ErrPayloadTooLarge = errors.New("payload too large")

	// ErrUnsupportedMediaType indicates that the Content-Type or Accept header is not supported.
	ErrUnsupportedMediaType = errors.New("unsupported media type")

	// ErrBindingFailed indicates a generic failure during the binding process.
	ErrBindingFailed = errors.New("binding failed")
)
