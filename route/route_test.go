package route

import (
	"context"
	"net/http"
	"testing"

	"github.com/cirius-go/miniapi"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoute(t *testing.T) {
	Convey("Given a new Route", t, func() {
		spec := Spec{
			Path:   "/test",
			Method: http.MethodPost,
			ID:     "testRoute",
		}

		handler := func(ctx context.Context, req *MockRequest) (*MockResponse, error) {
			return &MockResponse{}, nil
		}

		builder := NewBuilder(spec, handler)
		route := builder.Build(miniapi.BuildContext{})

		Convey("When accessing basic properties", func() {
			Convey("It should accurately report the Path", func() {
				So(route.Path(), ShouldEqual, "/test")
			})

			Convey("It should accurately report the Method", func() {
				So(route.Method(), ShouldEqual, http.MethodPost)
			})

			Convey("It should accurately report the Operation", func() {
				op := route.Operation()
				So(op.ID, ShouldEqual, "testRoute")
			})

			Convey("It should accurately report the HandlerFunc", func() {
				So(route.HandlerFunc(), ShouldNotBeNil)
			})
		})
	})
}
