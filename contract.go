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
	Default     bool
}

// TypedHandlerFunc represents a handler function in miniapi.
type TypedHandlerFunc[Rq, Rp any] func(ctx context.Context, rq *Rq) (*Rp, error)

// HandlerFunc represents the inner handler function in miniapi.
// The TypedHandler will be converted as HandlerFunc.
type HandlerFunc func(ctx Context)

// RouteModifier represents the function type for modifying a route.
type RouteModifier func(route Route) Route

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

// AuthEngine evaluates an incoming request against the required securities.
type AuthEngine interface {
	// Authenticate extracts the token/key, validates it, and either aborts the
	// request or injects the authenticated user's principal into the context.
	Authenticate() func(HandlerFunc) HandlerFunc
}

// Route represents a route in miniapi.
type Route interface {
	// Path returns the path of the route.
	Path() string
	// Method returns the method of route.
	Method() string
	// Operation returns the operation of the route.
	Operation() Operation
	// SetOperation sets the operation of the route.
	SetOperation(op Operation)
	// SetHandleFunc sets the handler function of the route.
	SetHandleFunc(fn HandlerFunc)
	// HandlerFunc returns the handler function of the route.
	HandlerFunc() HandlerFunc
	// ReqType returns the request reflection type.
	ReqType() reflect.Type
	// ResType returns the response reflection type.
	ResType() reflect.Type
	// AddModifiers adds the given modifiers to the route.
	AddModifiers(modifiers ...RouteModifier) Route
	// SetModifiers sets the modifiers of the route.
	SetModifiers(modifiers ...RouteModifier) Route
	// Modifiers returns the modifiers of the route.
	Modifiers() []RouteModifier
	// Middlewares returns the middlewares of the route.
	Middlewares() []Middleware
	// SetMiddlewares sets the middlewares of the route.
	SetMiddlewares(middlewares ...Middleware) Route
	// AddMiddlewares adds the middlewares to the route.
	AddMiddlewares(middlewares ...Middleware) Route
	// SetBinder sets the binder implementation for the route.
	SetBinder(b Binder) Route
	// Binder returns the binder implementation of the route.
	Binder() Binder
}

// Group represents the group of routes in miniapi.
type Group interface {
	// NewGroup creates a new group with the given path and adds it to the
	// current group.
	NewGroup(path string) Group
	// AddRoutes adds the given routes to the current group.
	AddRoutes(routes ...Route) Group
	// Path returns the path of the group.
	Path() string
	// FullPath returns the full path of the current group by concatenating the
	// paths of all parent groups.
	FullPath() string
	// Routes returns an iterator of all routes in the current group and
	// its subgroups.
	Routes() iter.Seq[Route]
	// Groups returns an iterator of all subgroups in the current group.
	Groups() iter.Seq[Group]
	// AddModifiers adds the given modifiers to the group.
	AddModifiers(modifiers ...RouteModifier) Group
	// Modifiers returns the modifiers of the group.
	Modifiers() []RouteModifier
	// WithSecurity overrides the global security for this group.
	WithSecurity(reqs ...SecurityRequirement) Group
	// Security returns the explicitly set OpenAPI security requirements, or nil.
	Security() []SecurityRequirement
	// AddMiddlewares adds the middlewares to the group.
	AddMiddlewares(middlewares ...Middleware) Group
	// Middlewares returns the middlewares of the group.
	Middlewares() []Middleware
	// SetMiddlewares sets the middlewares of the group.
	SetMiddlewares(middlewares ...Middleware) Group
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
	Attach(prefix string, group Group, adapter Adapter, moddifiers []RouteModifier, middlewares []Middleware) error
	Build(group Group) error
}

// Adapter represents the adapter for the mini API.
type Adapter interface {
	// AddRoute adds the given route to the adapter.
	AddRoute(prefix string, route Route)
	// Address returns the address of the adapter.
	Address() string
}
