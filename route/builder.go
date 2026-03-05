package route

import (
	"reflect"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/binder"
)

// Builder is a helper struct for building routes in miniapi.
type Builder struct {
	path         string
	method       string
	fn           miniapi.HandlerFuncBuilder
	operation    miniapi.Operation
	reqType      reflect.Type
	resType      reflect.Type
	middlewares  []miniapi.Middleware
	modifiers    []miniapi.Modifier
	binder       miniapi.Binder
	errorEncoder miniapi.ErrorEncoder
}

// Method implements miniapi.RouteBuilder.
func (h *Builder) Method() string {
	return h.method
}

// Path implements miniapi.RouteBuilder.
func (h *Builder) Path() string {
	return h.path
}

// SetMethod implements miniapi.RouteBuilder.
func (h *Builder) SetMethod(method string) miniapi.RouteBuilder {
	h.method = method
	return h
}

// SetPath implements miniapi.RouteBuilder.
func (h *Builder) SetPath(path string) miniapi.RouteBuilder {
	h.path = path
	return h
}

// HandlerFuncBuilder implements miniapi.RouteBuilder.
func (h *Builder) HandlerFuncBuilder() miniapi.HandlerFuncBuilder {
	return h.fn
}

// ErrorEncoder implements miniapi.RouteBuilder.
func (h *Builder) ErrorEncoder() miniapi.ErrorEncoder {
	return h.errorEncoder
}

// SetErrorEncoder implements miniapi.RouteBuilder.
func (h *Builder) SetErrorEncoder(e miniapi.ErrorEncoder) miniapi.RouteBuilder {
	h.errorEncoder = e
	return h
}

// ReqType implements miniapi.RouteBuilder.
func (h *Builder) ReqType() reflect.Type {
	return h.reqType
}

// ResType implements miniapi.RouteBuilder.
func (h *Builder) ResType() reflect.Type {
	return h.resType
}

// Binder implements miniapi.RouteBuilder.
func (h *Builder) Binder() miniapi.Binder {
	return h.binder
}

// Middlewares implements miniapi.RouteBuilder.
func (h *Builder) Middlewares() []miniapi.Middleware {
	return h.middlewares
}

// Modifiers implements miniapi.RouteBuilder.
func (h *Builder) Modifiers() []miniapi.Modifier {
	return h.modifiers
}

// Operation implements miniapi.RouteBuilder.
func (h *Builder) Operation() miniapi.Operation {
	return h.operation
}

// SetOperation implements miniapi.RouteBuilder.
func (h *Builder) SetOperation(op miniapi.Operation) miniapi.RouteBuilder {
	h.operation = op
	return h
}

// AddMiddlewares implements miniapi.RouteBuilder.
func (h *Builder) AddMiddlewares(middlewares ...miniapi.Middleware) miniapi.RouteBuilder {
	h.middlewares = append(h.middlewares, middlewares...)
	return h
}

// AddModifiers implements miniapi.RouteBuilder.
func (h *Builder) AddModifiers(modifiers ...miniapi.Modifier) miniapi.RouteBuilder {
	h.modifiers = append(h.modifiers, modifiers...)
	return h
}

// SetBinder implements miniapi.RouteBuilder.
func (h *Builder) SetBinder(binder miniapi.Binder) miniapi.RouteBuilder {
	h.binder = binder
	return h
}

// clone creates a copy of the builder to avoid modifying the original builder when applying modifiers.
func (h *Builder) clone() *Builder {
	newBuilder := *h
	newBuilder.middlewares = append([]miniapi.Middleware{}, h.middlewares...)
	newBuilder.modifiers = append([]miniapi.Modifier{}, h.modifiers...)
	return &newBuilder
}

// resolveModifiers applies the modifiers to the route builder and returns the
// modified route builder. It does pure transform builder, without modifying
// directly the original one.
func (h *Builder) resolveModifiers(extendedModifiers []miniapi.Modifier) miniapi.RouteBuilder {
	modifiers := append([]miniapi.Modifier{}, extendedModifiers...)
	modifiers = append(modifiers, h.modifiers...)
	var routeBuilder miniapi.RouteBuilder = h.clone()
	for _, mod := range modifiers {
		routeBuilder = mod(routeBuilder)
	}
	return routeBuilder
}

// resolveHandlerFunc resolves the handler function by applying the middlewares
// to the handler function built by the route builder.
func (h *Builder) resolveHandlerFunc(routeBuilder miniapi.RouteBuilder, extendedMiddlewares []miniapi.Middleware) miniapi.HandlerFunc {
	middlewares := append([]miniapi.Middleware{}, extendedMiddlewares...)
	middlewares = append(middlewares, routeBuilder.Middlewares()...)

	b := routeBuilder.Binder()
	if b == nil {
		b = binder.New(binder.DefaultConfig())
	}
	e := routeBuilder.ErrorEncoder()
	if e == nil {
		e = DefaultErrorEncoder
	}
	handlerFunc := routeBuilder.HandlerFuncBuilder()(b, e)
	handlerFunc = chain(middlewares, handlerFunc)
	return handlerFunc
}

// Build implements miniapi.RouteBuilder.
func (h *Builder) Build(extendedMiddlewares []miniapi.Middleware, extendedModifiers []miniapi.Modifier) miniapi.Route {
	routeBuilder := h.resolveModifiers(extendedModifiers)
	return &Route{
		path:      routeBuilder.Path(),
		method:    routeBuilder.Method(),
		operation: routeBuilder.Operation(),
		fn:        h.resolveHandlerFunc(routeBuilder, extendedMiddlewares),
	}
}

// Spec represents the specification for creating a new route.
type Spec struct {
	// These fields are required for creating a route.
	Path   string
	Method string

	// These fields are used for OpenAPI documentation.
	ID            string
	Tags          []string
	Summary       string
	Description   string
	Deprecated    bool
	Security      []miniapi.SecurityRequirement
	Extensions    map[string]any
	DefaultStatus int

	// ResponseTypes will override the default response types generated from the
	// response struct type for documenting purpose only.
	ResponseTypes map[int]miniapi.Response
	// Middlewares are the middlewares to be added to the route.
	Middlewares []miniapi.Middleware
	// Modifiers are the modifiers to be added to the route.
	Modifiers []miniapi.Modifier
	// Binder is used for binding request data and marshaling response data.
	Binder miniapi.Binder
	// ErrorEncoder is used for encoding errors.
	ErrorEncoder miniapi.ErrorEncoder
}

var _ miniapi.RouteBuilder = (*Builder)(nil)

// NewBuilder creates a new route builder with the given specification and typed handler function.
func NewBuilder[Rq, Rp any](spec Spec, typedFn miniapi.TypedHandlerFunc[Rq, Rp]) *Builder {
	responseTypes := make(map[int]miniapi.Response)
	if spec.ResponseTypes != nil {
		responseTypes = spec.ResponseTypes
	}

	r := &Builder{
		method:      spec.Method,
		path:        spec.Path,
		middlewares: make([]miniapi.Middleware, 0),
		modifiers:   make([]miniapi.Modifier, 0),
		operation: miniapi.Operation{
			ID:            spec.ID,
			Tags:          spec.Tags,
			Summary:       spec.Summary,
			Description:   spec.Description,
			Deprecated:    spec.Deprecated,
			Security:      spec.Security,
			DefaultStatus: spec.DefaultStatus,
			Extensions:    spec.Extensions,
			ResponseTypes: responseTypes,
		},
		reqType:      reflect.TypeOf(new(Rq)).Elem(),
		resType:      reflect.TypeOf(new(Rp)).Elem(),
		binder:       spec.Binder,
		errorEncoder: spec.ErrorEncoder,
		fn:           MakeHandlerFuncBuilder(typedFn),
	}

	return r
}

func chain(middlewares []miniapi.Middleware, final miniapi.HandlerFunc) miniapi.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] == nil {
			continue
		}
		final = middlewares[i](final)
	}
	return final
}
