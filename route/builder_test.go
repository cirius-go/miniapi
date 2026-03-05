package route

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/binder"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBuilder(t *testing.T) {
	Convey("Given a Route Builder", t, func() {
		spec := Spec{
			Path:          "/api/items",
			Method:        http.MethodGet,
			ID:            "getItems",
			Tags:          []string{"items", "public"},
			Summary:       "Get a list of items",
			Description:   "Returns a paginated list of items from the database.",
			Deprecated:    true,
			DefaultStatus: http.StatusAccepted,
			Extensions:    map[string]any{"x-custom": "value"},
			Security:      []miniapi.SecurityRequirement{{"oauth2": []string{"read:items"}}},
			ResponseTypes: map[int]miniapi.Response{
				http.StatusOK: JSONSchema[MockResponse]("Success response"),
			},
		}

		handler := func(ctx context.Context, req *MockRequest) (*MockResponse, error) {
			return &MockResponse{}, nil
		}

		builder := NewBuilder(spec, handler)

		Convey("It should accurately report the Path and Method", func() {
			So(builder.Path(), ShouldEqual, "/api/items")
			So(builder.Method(), ShouldEqual, http.MethodGet)
		})

		Convey("It should accurately report the Request/Response Types", func() {
			So(builder.ReqType(), ShouldEqual, reflect.TypeOf(MockRequest{}))
			So(builder.ResType(), ShouldEqual, reflect.TypeOf(MockResponse{}))
		})

		Convey("It should accurately store the OpenAPI metadata in the Operation object", func() {
			op := builder.Operation()
			So(op.ID, ShouldEqual, "getItems")
			So(op.Tags, ShouldResemble, []string{"items", "public"})
			So(op.Summary, ShouldEqual, "Get a list of items")
			So(op.Description, ShouldEqual, "Returns a paginated list of items from the database.")
			So(op.Deprecated, ShouldBeTrue)
			So(op.DefaultStatus, ShouldEqual, http.StatusAccepted)
			So(op.Extensions["x-custom"], ShouldEqual, "value")
			So(op.Security, ShouldResemble, []miniapi.SecurityRequirement{{"oauth2": []string{"read:items"}}})
			So(op.ResponseTypes, ShouldNotBeNil)
			So(op.ResponseTypes[http.StatusOK].Description, ShouldEqual, "Success response")
		})

		Convey("It should allow modifying Path and Method", func() {
			builder.SetPath("/new/path").SetMethod(http.MethodPut)
			So(builder.Path(), ShouldEqual, "/new/path")
			So(builder.Method(), ShouldEqual, http.MethodPut)
		})

		Convey("It should allow modifying Operation", func() {
			newOp := miniapi.Operation{ID: "newOperation"}
			builder.SetOperation(newOp)
			So(builder.Operation().ID, ShouldEqual, "newOperation")
		})

		Convey("It should correctly resolve defaults if missing from spec", func() {
			specDefault := Spec{
				Path:   "/default",
				Method: http.MethodGet,
			}
			builderDefault := NewBuilder(specDefault, handler)

			So(builderDefault.Binder(), ShouldBeNil)
			So(builderDefault.ErrorEncoder(), ShouldBeNil)
			So(builderDefault.Operation().ResponseTypes, ShouldBeEmpty)
			
			// Build applies the defaults
			route := builderDefault.Build(nil, nil)
			So(route, ShouldNotBeNil)
		})

		Convey("It should allow setting Binder and ErrorEncoder", func() {
			b := binder.New(binder.DefaultConfig())
			builder.SetBinder(b)
			So(builder.Binder(), ShouldEqual, b)

			e := func(c miniapi.Context, err error) {}
			builder.SetErrorEncoder(e)
			So(builder.ErrorEncoder(), ShouldNotBeNil)
		})

		Convey("It should allow adding and retrieving middlewares", func() {
			m1 := func(next miniapi.HandlerFunc) miniapi.HandlerFunc { return next }
			m2 := func(next miniapi.HandlerFunc) miniapi.HandlerFunc { return next }
			
			builder.AddMiddlewares(m1, m2)
			middlewares := builder.Middlewares()
			
			So(len(middlewares), ShouldEqual, 2)
		})

		Convey("It should correctly chain and execute middlewares in order", func() {
			executionOrder := []string{}
			
			m1 := func(next miniapi.HandlerFunc) miniapi.HandlerFunc {
				return func(c miniapi.Context) {
					executionOrder = append(executionOrder, "m1_start")
					next(c)
					executionOrder = append(executionOrder, "m1_end")
				}
			}
			
			m2 := func(next miniapi.HandlerFunc) miniapi.HandlerFunc {
				return func(c miniapi.Context) {
					executionOrder = append(executionOrder, "m2_start")
					next(c)
					executionOrder = append(executionOrder, "m2_end")
				}
			}
			
			finalHandler := func(c miniapi.Context) {
				executionOrder = append(executionOrder, "handler")
			}
			
			// Chain expects middlewares in the order they should be executed (outermost first)
			chained := chain([]miniapi.Middleware{m1, m2}, finalHandler)
			chained(nil) // Execute the chain
			
			expectedOrder := []string{"m1_start", "m2_start", "handler", "m2_end", "m1_end"}
			So(executionOrder, ShouldResemble, expectedOrder)
		})

		Convey("It should handle nil middlewares in the chain gracefully", func() {
			executionOrder := []string{}
			
			m1 := func(next miniapi.HandlerFunc) miniapi.HandlerFunc {
				return func(c miniapi.Context) {
					executionOrder = append(executionOrder, "m1")
					next(c)
				}
			}
			
			finalHandler := func(c miniapi.Context) {
				executionOrder = append(executionOrder, "handler")
			}
			
			chained := chain([]miniapi.Middleware{m1, nil}, finalHandler)
			chained(nil)
			
			So(executionOrder, ShouldResemble, []string{"m1", "handler"})
		})

		Convey("It should allow adding and resolving modifiers without mutating the original builder", func() {
			mod1 := func(b miniapi.RouteBuilder) miniapi.RouteBuilder {
				return b.SetPath("/modified/1")
			}
			mod2 := func(b miniapi.RouteBuilder) miniapi.RouteBuilder {
				return b.SetMethod(http.MethodDelete)
			}
			
			builder.AddModifiers(mod1)
			
			// resolveModifiers should apply existing and provided modifiers
			resolvedBuilder := builder.resolveModifiers([]miniapi.Modifier{mod2})
			
			So(resolvedBuilder.Path(), ShouldEqual, "/modified/1")
			So(resolvedBuilder.Method(), ShouldEqual, http.MethodDelete)
			
			// Original builder should not be mutated
			So(builder.Path(), ShouldEqual, "/api/items")
			So(builder.Method(), ShouldEqual, http.MethodGet)
			So(len(builder.Modifiers()), ShouldEqual, 1)
		})
	})
}
