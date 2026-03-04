package openapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/cirius-go/miniapi/group"
	"github.com/cirius-go/miniapi/openapi"
	"github.com/cirius-go/miniapi/route"
	"github.com/getkin/kin-openapi/openapi3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRouteSecurityOverrides(t *testing.T) {
	Convey("Given an API with global security", t, func() {
		g := group.New("/api")

		type Empty struct{}
		r1 := route.New[Empty, Empty](route.Spec{Path: "/public", Method: "GET"}, func(ctx context.Context, req *Empty) (*Empty, error) {
			return &Empty{}, nil
		})
		r1.WithSecurity(openapi.PublicSecurity()...)

		r2 := route.New[Empty, Empty](route.Spec{Path: "/custom", Method: "GET"}, func(ctx context.Context, req *Empty) (*Empty, error) {
			return &Empty{}, nil
		})
		r2.WithSecurity(
			openapi3.SecurityRequirement{"ApiKeyAuth": []string{}},
			openapi3.SecurityRequirement{"BearerAuth": []string{"admin"}},
		)

		r3 := route.New[Empty, Empty](route.Spec{Path: "/strict", Method: "GET"}, func(ctx context.Context, req *Empty) (*Empty, error) {
			return &Empty{}, nil
		})
		r3.WithSecurity(openapi3.SecurityRequirement{
			"ApiKeyAuth": []string{},
			"BearerAuth": []string{"admin"},
		})

		g.AddRoutes(r1, r2, r3)

		adapter := &dummyAdapter{}

		cfg := openapi.BuildConfig{}
		cfg.Security = openapi3.SecurityRequirements{
			{"GlobalAuth": []string{}},
		}

		err := openapi.Build(g, adapter, cfg)
		So(err, ShouldBeNil)

		docsRoute, ok := adapter.routes["/docs/openapi.json"]
		So(ok, ShouldBeTrue)

		ctx := &mockReqContext{route: docsRoute}
		docsRoute.HandlerFunc()(ctx)

		out := ctx.ResponseBodyWriter().(*bytes.Buffer).Bytes()

		var doc openapi3.T
		err = json.Unmarshal(out, &doc)
		So(err, ShouldBeNil)

		Convey("Public route should have empty security array", func() {
			op := doc.Paths.Value("/api/public").Get
			So(op.Security, ShouldNotBeNil)
			So(len(*op.Security), ShouldEqual, 0)
		})

		Convey("Custom route should have OR logic", func() {
			op := doc.Paths.Value("/api/custom").Get
			So(op.Security, ShouldNotBeNil)
			So(len(*op.Security), ShouldEqual, 2)
			So((*op.Security)[0], ShouldContainKey, "ApiKeyAuth")
			So((*op.Security)[1], ShouldContainKey, "BearerAuth")
			So((*op.Security)[1]["BearerAuth"], ShouldResemble, []string{"admin"})
		})

		Convey("Strict route should have AND logic", func() {
			op := doc.Paths.Value("/api/strict").Get
			So(op.Security, ShouldNotBeNil)
			So(len(*op.Security), ShouldEqual, 1)
			req := (*op.Security)[0]
			So(req, ShouldContainKey, "ApiKeyAuth")
			So(req, ShouldContainKey, "BearerAuth")
			So(req["BearerAuth"], ShouldResemble, []string{"admin"})
		})
	})
}
