package openapi_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/group"
	"github.com/cirius-go/miniapi/openapi"
	"github.com/cirius-go/miniapi/route"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

type mockAuthEngine struct {
	Called bool
	Secs   openapi3.SecurityRequirements
}

func (m *mockAuthEngine) Authenticate(secs openapi3.SecurityRequirements) func(miniapi.HandlerFunc) miniapi.HandlerFunc {
	return func(next miniapi.HandlerFunc) miniapi.HandlerFunc {
		return func(ctx miniapi.Context) {
			m.Called = true
			m.Secs = secs
			if ctx.RequestHeader("Authorization") == "" {
				ctx.SetResponseStatus(http.StatusUnauthorized)
				return
			}
			next(ctx)
		}
	}
}

type mockReqContext struct {
	miniapi.Context
	header string
	status int
	route  miniapi.Route
	resBuf *bytes.Buffer
}

func (m *mockReqContext) Route() miniapi.Route           { return m.route }
func (m *mockReqContext) RequestHeader(k string) string  { return m.header }
func (m *mockReqContext) SetResponseStatus(s int)        { m.status = s }
func (m *mockReqContext) SetResponseHeader(k, v string)  {}
func (m *mockReqContext) ResponseHeader(k string) string { return "" }
func (m *mockReqContext) ResponseStatus() int            { return m.status }
func (m *mockReqContext) ResponseBodyWriter() io.Writer {
	if m.resBuf == nil {
		m.resBuf = &bytes.Buffer{}
	}
	return m.resBuf
}
func (m *mockReqContext) RequestContext() context.Context { return context.Background() }

type mockBinder struct{}

func (m *mockBinder) BindRequest(ctx miniapi.Context, req any) error     { return nil }
func (m *mockBinder) MarshalResponse(ctx miniapi.Context, res any) error { return nil }

func TestBuilder_AuthEngineMiddleware(t *testing.T) {
	g := group.New("/api")

	type Empty struct{}
	r := route.New[Empty, Empty](route.Spec{Path: "/secure", Method: "GET"}, func(ctx context.Context, req *Empty) (*Empty, error) {
		return &Empty{}, nil
	})
	r.SetBinder(&mockBinder{})

	r.WithSecurity(openapi3.SecurityRequirement{"BearerAuth": []string{"admin"}})
	g.AddRoutes(r)

	engine := &mockAuthEngine{}
	adapter := &dummyAdapter{}
	err := openapi.Build(g, adapter, openapi.BuildConfig{
		AuthEngine: engine,
	})

	assert.NoError(t, err)

	secureRoute, ok := adapter.routes["/api/secure"]
	assert.True(t, ok)

	handler := secureRoute.HandlerFunc()

	// 1. Missing Auth
	ctx1 := &mockReqContext{header: "", route: secureRoute}
	handler(ctx1)
	assert.True(t, engine.Called)
	assert.Equal(t, http.StatusUnauthorized, ctx1.status)
	assert.Contains(t, engine.Secs[0], "BearerAuth")

	// 2. Valid Auth
	ctx2 := &mockReqContext{header: "Bearer token", route: secureRoute}
	handler(ctx2)
	assert.Equal(t, http.StatusOK, ctx2.status) // Expect 200 OK
}

func TestBuilder_AuthEngineOpenAPI(t *testing.T) {
	g := group.New("/api")

	type Empty struct{}
	r := route.New[Empty, Empty](route.Spec{Path: "/secure", Method: "GET"}, func(ctx context.Context, req *Empty) (*Empty, error) {
		return &Empty{}, nil
	})

	r.WithSecurity(openapi3.SecurityRequirement{"BearerAuth": []string{"admin"}})
	g.AddRoutes(r)

	engine := &mockAuthEngine{}
	adapter := &dummyAdapter{}
	err := openapi.Build(g, adapter, openapi.BuildConfig{
		AuthEngine: engine,
	})

	assert.NoError(t, err)

	docsRoute, ok := adapter.routes["/docs/openapi.json"]
	assert.True(t, ok)

	handler := docsRoute.HandlerFunc()
	ctx := &mockReqContext{header: "", route: docsRoute}
	handler(ctx)

	out := ctx.ResponseBodyWriter().(*bytes.Buffer).String()

	// Test if security scopes are properly serialized into the document
	assert.Contains(t, out, `"security":[{"BearerAuth":["admin"]}]`)
}
