package miniapi

import (
	"context"
	"crypto/tls"
	"io"
	"iter"
	"mime/multipart"
	"net/url"
)

// TypedHandlerFunc represents a handler function in miniapi.
type TypedHandlerFunc[Rq, Rp any] func(ctx context.Context, rq *Rq) (*Rp, error)

// HandlerFunc represents the inner handler function in miniapi.
// The TypedHandler will be converted as HandlerFunc.
type HandlerFunc func(ctx Context)

// Route represents a route in miniapi.
type Route interface {
	// Path returns the path of the route.
	Path() string
	// Method returns the method of the route.
	Method() string
}

// Context represents the context of a request in miniapi.
type Context interface {
	// Route returns the route definition.
	Route() Route

	// RequestContext returns the context of the request.
	RequestContext() context.Context
	// RequestMethod returns the method of the request.
	RequestMethod() string
	// RequestHost returns the host of the request.
	RequestHost() string
	// RemoteAddress returns the remote address of the request.
	RemoteAddress() string
	// RequestURL returns the URL of the request.
	RequestURL() url.URL
	// RequestParam returns the value of the path parameter with the given name.
	RequestParam(name string) string
	// RequestQuery returns the value of the query parameter with the
	// given name.
	RequestQuery(name string) string
	// RequestHeader returns the value of the header with the given name.
	RequestHeader(name string) string
	// RequestTLS returns the connection state of the TLS connection.
	RequestTLS() *tls.ConnectionState
	// IterRequestParam returns an iterator of all path parameters.
	IterRequestHeader() iter.Seq2[string, string]
	// RequestBodyReader returns the reader of the request body.
	RequestBodyReader() io.ReadCloser
	// RequestMultipartForm returns the multipart form of the request.
	RequestMultipartForm(maxMemory int64) (*multipart.Form, error)
	// SetResponseStatus sets the status code of the response.
	SetResponseStatus(status int)
	// ResponseStatus returns the status code of the response.
	ResponseStatus() int
	// SetResponseHeader sets the header of the response.
	SetResponseHeader(name, value string)
	// AppendHeader appends the header of the response.
	AppendResponseHeader(name, value string)
	// ResponseBodyWriter returns the writer of the response body.
	ResponseBodyWriter() io.Writer
}
