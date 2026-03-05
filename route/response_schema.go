package route

import (
	"reflect"

	"github.com/cirius-go/miniapi"
)

// JSONSchema creates a miniapi.Response with the schema of type T and an empty description.
func JSONSchema[T any](desc string) miniapi.Response {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return miniapi.Response{
		ContentType: "application/json",
		Description: desc,
		Schema:      t,
	}
}

// ProblemSchema creates a miniapi.Response with the schema of type T and an
// empty description, and sets the content type to "application/problem+json".
func ProblemSchema[T any](desc string) miniapi.Response {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return miniapi.Response{
		ContentType: "application/problem+json",
		Description: desc,
		Schema:      t,
	}
}

// NoContentSchema creates a miniapi.Response with no content and an empty description.
func NoContentSchema(desc string) miniapi.Response {
	return miniapi.Response{
		Description: desc,
		Schema:      nil,
	}
}
