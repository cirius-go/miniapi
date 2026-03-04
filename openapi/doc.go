package openapi

import (
	"net/http"

	"github.com/cirius-go/miniapi"
)

// BuildDocsHandler builds a handler function that serves the OpenAPI specification document.
func BuildDocsHandler(docData []byte) miniapi.HandlerFunc {
	return func(ctx miniapi.Context) {
		ctx.SetResponseHeader("Content-Type", "application/json")
		ctx.SetResponseStatus(http.StatusOK)
		_, _ = ctx.ResponseBodyWriter().Write(docData)
	}
}
