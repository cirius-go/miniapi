package route

import "github.com/cirius-go/miniapi"

// Operation represents an operation in a route.
type Operation struct {
	ID          string
	Tags        []string
	Summary     string
	Description string
}

// Route represents a route in miniapi.
type Route struct {
	Path        string
	HandlerFunc miniapi.HandlerFunc
	Operation   *Operation
}

// NewRoute creates a new route with path.
func NewRoute[Rq, Rp any](path string, handler miniapi.TypedHandlerFunc[Rq, Rp], operation *Operation) *Route {
	return &Route{
		Path: path,
		HandlerFunc: func(c miniapi.Context) {
			var (
				ctx = c.RequestContext()
				req = new(Rq)
			)

			// TODO: handle response, error
			_, _ = handler(ctx, req)
		},
		Operation: operation,
	}
}
