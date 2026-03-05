package route

import (
	"github.com/cirius-go/miniapi"
)

// Route represents a route in miniapi.
type Route struct {
	path      string
	method    string
	operation miniapi.Operation
	fn        miniapi.HandlerFunc
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

// HandlerFunc implements miniapi.Route.
func (r *Route) HandlerFunc() miniapi.HandlerFunc {
	return r.fn
}

var _ miniapi.Route = (*Route)(nil)
