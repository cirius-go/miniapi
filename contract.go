package miniapi

import (
	"context"
	"crypto/tls"
	"io"
	"iter"
	"mime/multipart"
	"net/url"
	"reflect"
)

// Response represents the response of a route in miniapi.
type Response struct {
	ContentType string
	Description string
	Schema      reflect.Type
	NoContent   bool
}

// TypedHandlerFunc represents a handler function in miniapi.
type TypedHandlerFunc[Rq, Rp any] func(ctx context.Context, rq *Rq) (*Rp, error)

// HandlerFunc represents the inner handler function in miniapi.
// The TypedHandler will be converted as HandlerFunc.
type HandlerFunc func(ctx Context)

// Middleware represents the function type for modifying a handler function.
type Middleware func(HandlerFunc) HandlerFunc

// SecurityRequirement represents the security requirement of a route in miniapi.
// scheme:scopes[]
type SecurityRequirement map[string][]string

// Operation represents the operation of a route in miniapi.
type Operation struct {
	ID          string
	Summary     string
	Description string

	ResponseTypes map[int]Response

	Tags       []string
	Security   []SecurityRequirement
	Extensions map[string]any

	Deprecated    bool
	DefaultStatus int
}

// Binder defines the interface for binding request data to structs
// and marshaling response structs to the HTTP output.
type Binder interface {
	// BindRequest parses path, query, header, and body parameters into the given req struct.
	BindRequest(ctx Context, req any) error

	// MarshalResponse marshals the given res struct into the HTTP response writer.
	MarshalResponse(ctx Context, res any) error
}

// ErrorEncoder represents the function type for encoding an error into the
// response.
type ErrorEncoder func(Context, error)

// HandlerFuncBuilder is a function type that builds a HandlerFunc using a
// Binder and an ErrorEncoder.
type HandlerFuncBuilder func(b Binder, e ErrorEncoder) HandlerFunc

// Modifier represents the function type for modifying a route builder.
type Modifier func(route RouteBuilder) RouteBuilder

// RouteBuilder is used for building a route in miniapi.
type RouteBuilder interface {
	// Path returns the path of the route builder.
	Path() string
	// SetPath sets the path of the route builder.
	SetPath(path string) RouteBuilder
	// Method returns the method of the route builder.
	Method() string
	// SetMethod sets the method of the route builder.
	SetMethod(method string) RouteBuilder
	// Modifiers returns the modifiers of the route builder.
	Modifiers() []Modifier
	// AddModifiers adds the given modifiers to the route builder.
	AddModifiers(modifiers ...Modifier) RouteBuilder
	// Middlewares returns the middlewares of the route builder.
	Middlewares() []Middleware
	// AddMiddlewares adds the given middlewares to the route builder.
	AddMiddlewares(middlewares ...Middleware) RouteBuilder
	// Binder returns the binder implementation of the route builder.
	Binder() Binder
	// SetBinder sets the binder implementation for the route builder.
	SetBinder(b Binder) RouteBuilder
	// Operation returns the operation of the route builder.
	Operation() Operation
	// SetOperation sets the operation of the route builder.
	SetOperation(op Operation) RouteBuilder
	// ReqType returns the request reflection type of the route builder.
	ReqType() reflect.Type
	// ResType returns the response reflection type of the route builder.
	ResType() reflect.Type
	// ErrorEncoder returns the error encoder of the route builder.
	ErrorEncoder() ErrorEncoder
	// SetErrorEncoder sets the error encoder for the route builder.
	SetErrorEncoder(e ErrorEncoder) RouteBuilder
	// HandlerFuncBuilder returns the handler function builder of the route builder.
	HandlerFuncBuilder() HandlerFuncBuilder

	// Build builds the route with the given specification and typed handler function.
	Build(ctx BuildContext) Route
}

// Route represents a route in miniapi.
type Route interface {
	// Path returns the path of the route.
	Path() string
	// Method returns the method of route.
	Method() string
	// Operation returns the operation of the route.
	Operation() Operation
	// HandlerFunc returns the handler function of the route.
	HandlerFunc() HandlerFunc
}

// Group represents the group of routes in miniapi.
type Group interface {
	// AddGroups adds the given groups to the current group.
	AddGroups(groups ...Group) Group
	// AddRoutes adds the given routes to the current group.
	AddRoutes(routes ...RouteBuilder) Group
	// Prefix returns the prefix of the group.
	Prefix() string
	// SetPrefix sets the prefix of the group.
	SetPrefix(prefix string) Group
	// AddModifiers adds the given modifiers to the group.
	AddModifiers(modifiers ...Modifier) Group
	// Modifiers returns the modifiers of the group.
	Modifiers() []Modifier
	// AddMiddlewares adds the middlewares to the group.
	AddMiddlewares(middlewares ...Middleware) Group
	// Middlewares returns the middlewares of the group.
	Middlewares() []Middleware
	// Binder returns the binder implementation of the group.
	Binder() Binder
	// SetBinder sets the binder implementation for the group.
	SetBinder(b Binder) Group
	// ErrorEncoder returns the error encoder of the group.
	ErrorEncoder() ErrorEncoder
	// SetErrorEncoder sets the error encoder for the group.
	SetErrorEncoder(e ErrorEncoder) Group
	// NewGroup creates a new child group with the given prefix and adds it to the current group.
	NewGroup(prefix string) Group
	// Build builds the group and returns the list of routes in the group.
	Build(BuildContext) []Route
}

// BuildContext represents the context for building the mini API. It can be
// used to store any information needed during the build process.
type BuildContext struct {
	Prefix       string
	ErrorEncoder ErrorEncoder
	Binder       Binder
	Middlewares  []Middleware
	Modifiers    []Modifier
}

// Clone clones the build context to avoid modifying the original one when applying modifiers.
func (ctx BuildContext) Clone() BuildContext {
	return BuildContext{
		Prefix:       ctx.Prefix,
		ErrorEncoder: ctx.ErrorEncoder,
		Binder:       ctx.Binder,
		Middlewares:  append([]Middleware{}, ctx.Middlewares...),
		Modifiers:    append([]Modifier{}, ctx.Modifiers...),
	}
}

// Context represents the context of a request in miniapi.
type Context interface {
	// Route returns the route definition.
	Route() Route

	// RequestContext returns the context of the request.
	RequestContext() context.Context
	// SetRequestContext sets the context of the request.
	SetRequestContext(ctx context.Context)
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
	// ResponseHeader returns the value of the response header with the given name.
	ResponseHeader(name string) string
	// SetResponseHeader sets the header of the response.
	SetResponseHeader(name, value string)
	// AppendHeader appends the header of the response.
	AppendResponseHeader(name, value string)
	// ResponseBodyWriter returns the writer of the response body.
	ResponseBodyWriter() io.Writer
}

// OpenAPI represents the OpenAPI specification of the mini API.
type OpenAPI interface {
	// Attach attaches the OpenAPI specification to the given group and adapter.
	Attach(prefix string, group Group, adapter Adapter, moddifiers []Modifier, middlewares []Middleware) error
	Build(group Group) error
}

// Adapter represents the adapter for the mini API.
type Adapter interface {
	// AddRoute adds the given route to the adapter.
	AddRoute(prefix string, route Route)
	// Address returns the address of the adapter.
	Address() string
}
