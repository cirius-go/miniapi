package group

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestGroup_WithSecurity(t *testing.T) {
	g := New("/api")
	g.WithSecurity(
		openapi3.SecurityRequirement{"BearerAuth": []string{"read", "write"}},
		openapi3.SecurityRequirement{"ApiKeyAuth": []string{}},
	)

	sec := g.Security()
	assert.NotNil(t, sec)
	assert.Len(t, *sec, 2)
	assert.Contains(t, (*sec)[0], "BearerAuth")
	assert.Equal(t, []string{"read", "write"}, (*sec)[0]["BearerAuth"])
	assert.Contains(t, (*sec)[1], "ApiKeyAuth")
	assert.Equal(t, []string{}, (*sec)[1]["ApiKeyAuth"])
}
