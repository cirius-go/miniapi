package openapi_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/group"
	"github.com/cirius-go/miniapi/openapi"
	"github.com/cirius-go/miniapi/route"
	"github.com/stretchr/testify/assert"
)

type dummyAdapter struct {
	routes map[string]miniapi.Route
}

func (a *dummyAdapter) AddRoute(prefix string, r miniapi.Route) {
	if a.routes == nil {
		a.routes = make(map[string]miniapi.Route)
	}
	a.routes[prefix+r.Path()] = r
}

type mockContext struct {
	miniapi.Context
	resBuf *bytes.Buffer
}

func (m *mockContext) SetResponseStatus(s int)         {}
func (m *mockContext) SetResponseHeader(k, v string)   {}
func (m *mockContext) ResponseBodyWriter() io.Writer   { return m.resBuf }
func (m *mockContext) RequestContext() context.Context { return context.Background() }

func TestDuplicateOpenAPI(t *testing.T) {
	g := group.New("/api")

	type Empty struct{}
	r := route.New[Empty, Empty](route.Spec{Path: "/hello", Method: "GET", ID: "hello", Summary: "hello"}, func(ctx context.Context, req *Empty) (*Empty, error) {
		return &Empty{}, nil
	})

	// Add a modifier that appends a tag. If buildPaths runs multiple times, tags will duplicate.
	r.AddModifiers(func(rt miniapi.Route) miniapi.Route {
		op := rt.Operation()
		op.Tags = append(op.Tags, "AddedTag")
		rt.SetOperation(op)
		return rt
	})

	g.AddRoutes(r)

	adapter := &dummyAdapter{}
	err := openapi.Build(g, adapter, openapi.BuildConfig{})
	assert.NoError(t, err)

	docsRoute, ok := adapter.routes["/docs/openapi.json"]
	assert.True(t, ok)

	handler := docsRoute.HandlerFunc()

	var firstOutput string

	for i := 0; i < 3; i++ {
		mctx := &mockContext{resBuf: &bytes.Buffer{}}
		handler(mctx)

		out := mctx.resBuf.String()
		if i == 0 {
			firstOutput = out
		} else {
			assert.Equal(t, firstOutput, out, "Output changed on request %d", i+1)
		}
	}
}

func TestDuplicateRouteModifiers(t *testing.T) {
	g := group.New("/api")

	type Empty struct{}
	r := route.New[Empty, Empty](route.Spec{Path: "/hello", Method: "GET", ID: "hello", Summary: "hello"}, func(ctx context.Context, req *Empty) (*Empty, error) {
		return &Empty{}, nil
	})

	// Add a modifier that appends a tag
	r.AddModifiers(func(rt miniapi.Route) miniapi.Route {
		op := rt.Operation()
		op.Tags = append(op.Tags, "AddedTag")
		rt.SetOperation(op)
		return rt
	})

	g.AddRoutes(r)

	adapter := &dummyAdapter{}
	err := openapi.Build(g, adapter, openapi.BuildConfig{})
	assert.NoError(t, err)

	docsRoute, ok := adapter.routes["/docs/openapi.json"]
	assert.True(t, ok)

	handler := docsRoute.HandlerFunc()
	mctx := &mockContext{resBuf: &bytes.Buffer{}}
	handler(mctx)

	out := mctx.resBuf.String()

	// Verify that the tag "AddedTag" appears exactly once in the generated JSON
	// And not multiple times like "AddedTag","AddedTag"
	assert.Contains(t, out, `"tags":["AddedTag"]`)
	assert.NotContains(t, out, `"tags":["AddedTag","AddedTag"]`)
}

func BenchmarkBuild(b *testing.B) {
	g := group.New("/api")
	type Empty struct{}
	r := route.New[Empty, Empty](route.Spec{Path: "/hello", Method: "GET", ID: "hello", Summary: "hello"}, func(ctx context.Context, req *Empty) (*Empty, error) {
		return &Empty{}, nil
	})
	r.AddModifiers(func(rt miniapi.Route) miniapi.Route {
		op := rt.Operation()
		op.Tags = append(op.Tags, "AddedTag")
		rt.SetOperation(op)
		return rt
	})
	g.AddRoutes(r)
	adapter := &dummyAdapter{}
	cfg := openapi.BuildConfig{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = openapi.Build(g, adapter, cfg)
	}
}
