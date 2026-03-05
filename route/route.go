package route

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/binder"
)

// Route represents a route in miniapi.
type Route struct {
	path        string
	method      string
	fn          miniapi.HandlerFunc
	operation   miniapi.Operation
	reqType     reflect.Type
	resType     reflect.Type
	middlewares []miniapi.Middleware
	modifiers   []miniapi.RouteModifier
	binder      miniapi.Binder
}

// Path implements miniapi.Route.
func (r *Route) Path() string {
	return r.path
}

// Method implements miniapi.Route.
func (r *Route) Method() string {
	return r.method
}

// Operation implements miniapi.Route.
func (r *Route) Operation() miniapi.Operation {
	return r.operation
}

// SetOperation implements miniapi.Route.
func (r *Route) SetOperation(op miniapi.Operation) {
	r.operation = op
}

// SetHandleFunc implements miniapi.Route.
func (r *Route) SetHandleFunc(fn miniapi.HandlerFunc) {
	r.fn = fn
}

// SetModifiers implements miniapi.Route.
func (r *Route) SetModifiers(modifiers ...miniapi.RouteModifier) miniapi.Route {
	r.modifiers = modifiers
	return r
}

// AddMiddlewares implements miniapi.Route.
func (r *Route) AddMiddlewares(middlewares ...miniapi.Middleware) miniapi.Route {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// SetMiddlewares implements miniapi.Route.
func (r *Route) SetMiddlewares(middlewares ...miniapi.Middleware) miniapi.Route {
	r.middlewares = middlewares
	return r
}

// Middlewares implements miniapi.Route.
func (r *Route) Middlewares() []miniapi.Middleware {
	return r.middlewares
}

// AddModifiers implements miniapi.Route.
func (r *Route) AddModifiers(modifiers ...miniapi.RouteModifier) miniapi.Route {
	r.modifiers = append(r.modifiers, modifiers...)
	return r
}

// Modifiers implements miniapi.Route.
func (r *Route) Modifiers() []miniapi.RouteModifier {
	return r.modifiers
}

// HandlerFunc implements miniapi.Route.
func (r *Route) HandlerFunc() miniapi.HandlerFunc {
	return r.fn
}

// ReqType implements miniapi.Route.
func (r *Route) ReqType() reflect.Type {
	return r.reqType
}

// ResType implements miniapi.Route.
func (r *Route) ResType() reflect.Type {
	return r.resType
}

// SetBinder implements miniapi.Route.
func (r *Route) SetBinder(b miniapi.Binder) miniapi.Route {
	r.binder = b
	return r
}

// Binder implements miniapi.Route.
func (r *Route) Binder() miniapi.Binder {
	return r.binder
}

var _ miniapi.Route = (*Route)(nil)

// Spec represents the specification of a route, which can be used to
// generate OpenAPI documentation.
type Spec struct {
	Path          string
	Method        string
	DefaultStatus int

	ID          string
	Tags        []string
	Summary     string
	Description string
	Deprecated  bool

	Security []miniapi.SecurityRequirement
	// ResponseTypes will override the default response types generated from the
	// response struct type for documenting purpose only.
	ResponseTypes map[int]miniapi.Response
}

// New creates a new route with path.
func New[Rq, Rp any](spec Spec, typedFn miniapi.TypedHandlerFunc[Rq, Rp]) *Route {
	responseTypes := make(map[int]miniapi.Response)
	if spec.ResponseTypes != nil {
		responseTypes = spec.ResponseTypes
	}
	defaultStatus := spec.DefaultStatus
	if defaultStatus == 0 {
		if len(responseTypes) > 0 {
			for status := range responseTypes {
				defaultStatus = status
				break
			}
		} else {
			defaultStatus = http.StatusOK
		}
	}
	r := &Route{
		method:      spec.Method,
		path:        spec.Path,
		middlewares: make([]miniapi.Middleware, 0),
		modifiers:   make([]miniapi.RouteModifier, 0),
		operation: miniapi.Operation{
			ID:            spec.ID,
			Tags:          spec.Tags,
			Summary:       spec.Summary,
			Description:   spec.Description,
			Deprecated:    spec.Deprecated,
			Security:      spec.Security,
			DefaultStatus: spec.DefaultStatus,
			ResponseTypes: responseTypes,
		},
		reqType: reflect.TypeOf(new(Rq)).Elem(),
		resType: reflect.TypeOf(new(Rp)).Elem(),
		binder:  binder.New(binder.DefaultConfig()),
	}

	r.fn = func(c miniapi.Context) {
		var (
			ctx = c.RequestContext()
			req = new(Rq)
		)

		// Decode request
		if err := c.Route().Binder().BindRequest(c, req); err != nil {
			if c.ResponseHeader("Content-Type") == "" {
				c.SetResponseHeader("Content-Type", "application/problem+json")
			}
			c.SetResponseStatus(http.StatusBadRequest)
			errRes := struct {
				Error string `json:"error"`
			}{Error: "failed to decode request: " + err.Error()}
			_ = json.NewEncoder(c.ResponseBodyWriter()).Encode(errRes)
			return
		}

		// Call handler
		res, err := typedFn(ctx, req)
		if err != nil {
			if c.ResponseHeader("Content-Type") == "" {
				c.SetResponseHeader("Content-Type", "application/problem+json")
			}
			if c.ResponseStatus() == 0 {
				c.SetResponseStatus(http.StatusInternalServerError)
			}
			errRes := struct {
				Error string `json:"error"`
			}{Error: err.Error()}
			_ = json.NewEncoder(c.ResponseBodyWriter()).Encode(errRes)
			return
		}

		// Success
		if c.ResponseStatus() == 0 {
			c.SetResponseStatus(http.StatusOK)
		}
		_ = c.Route().Binder().MarshalResponse(c, res)
	}

	return r
}

// NewWithHandlerFunc creates a new route with path and handler function.
func NewWithHandlerFunc(spec Spec, fn miniapi.HandlerFunc) *Route {
	responseTypes := make(map[int]miniapi.Response)
	if spec.ResponseTypes != nil {
		responseTypes = spec.ResponseTypes
	}
	return &Route{
		method:      spec.Method,
		path:        spec.Path,
		fn:          fn,
		middlewares: make([]miniapi.Middleware, 0),
		modifiers:   make([]miniapi.RouteModifier, 0),
		operation: miniapi.Operation{
			ID:            spec.ID,
			Tags:          spec.Tags,
			Summary:       spec.Summary,
			Description:   spec.Description,
			Deprecated:    spec.Deprecated,
			Security:      spec.Security,
			ResponseTypes: responseTypes,
		},
		reqType: nil,
		resType: nil,
		binder:  binder.New(binder.DefaultConfig()),
	}
}
