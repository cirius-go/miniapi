package openapi

import "github.com/getkin/kin-openapi/openapi3"

// PublicSecurity returns an empty SecurityRequirements slice,
// which explicitly marks a route as public, overriding any
// global security requirements.
func PublicSecurity() openapi3.SecurityRequirements {
	return []openapi3.SecurityRequirement{}
}
