package route

import (
	"reflect"

	"github.com/cirius-go/miniapi"
)

// JSONOf creates a miniapi.Response with the schema of type T and an empty description.
func JSONOf[T any](desc string, ds ...bool) miniapi.Response {
	t := reflect.TypeOf((*T)(nil)).Elem()
	d := false
	if len(ds) > 0 && ds[0] {
		d = true
	}
	return miniapi.Response{
		ContentType: "application/json",
		Description: desc,
		Schema:      t,
		Default:     d,
	}
}

// ProblemJSONOf creates a miniapi.Response with the schema of type T and an
// empty description, and sets the content type to "application/problem+json".
func ProblemJSONOf[T any](desc string, ds ...bool) miniapi.Response {
	t := reflect.TypeOf((*T)(nil)).Elem()
	d := false
	if len(ds) > 0 && ds[0] {
		d = true
	}
	return miniapi.Response{
		ContentType: "application/problem+json",
		Description: desc,
		Schema:      t,
		Default:     d,
	}
}

// NoContent creates a miniapi.Response with no content and an empty description.
func NoContent(desc string, ds ...bool) miniapi.Response {
	d := false
	if len(ds) > 0 && ds[0] {
		d = true
	}
	return miniapi.Response{
		Description: desc,
		Schema:      nil,
		Default:     d,
	}
}
