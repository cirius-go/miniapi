package route_test

import (
	"context"
	"testing"

	"github.com/cirius-go/miniapi/route"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestRoute_WithSecurity(t *testing.T) {
	type Empty struct{}
	r := route.New[Empty, Empty](route.Spec{Path: "/test", Method: "GET"}, func(ctx context.Context, req *Empty) (*Empty, error) {
		return &Empty{}, nil
	})

	r.WithSecurity(
		openapi3.SecurityRequirement{"BearerAuth": []string{"admin"}},
		openapi3.SecurityRequirement{"BasicAuth": []string{}},
	)

	sec := r.Security()
	assert.NotNil(t, sec)
	assert.Len(t, *sec, 2)
	assert.Contains(t, (*sec)[0], "BearerAuth")
	assert.Equal(t, []string{"admin"}, (*sec)[0]["BearerAuth"])
	assert.Contains(t, (*sec)[1], "BasicAuth")
	assert.Equal(t, []string{}, (*sec)[1]["BasicAuth"])
}
