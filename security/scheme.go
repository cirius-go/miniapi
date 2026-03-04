package security

// NewRequirement creates a new SecurityRequirement with the given scheme and scopes.
func NewRequirement(scheme string, scopes ...string) map[string][]string {
	return map[string][]string{
		scheme: scopes,
	}
}
