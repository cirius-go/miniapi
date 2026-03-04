package openapi

import (
	"fmt"
	"net/http"

	"github.com/cirius-go/miniapi"
)

const redocTemplate = `<!DOCTYPE html>
<html>
  <head>
    <title>Redoc API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
      body { margin: 0; padding: 0; }
    </style>
  </head>
  <body>
    <redoc spec-url="%s"></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"> </script>
  </body>
</html>`

const scalarTemplate = `<!doctype html>
<html>
  <head>
    <title>Scalar API Reference</title>
    <meta charset="utf-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script
      id="api-reference"
      data-url="%s"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

// BuildUIHandler returns a handler that serves the interactive OpenAPI documentation UI.
func BuildUIHandler(docPath string, lib UILib) miniapi.HandlerFunc {
	return func(ctx miniapi.Context) {
		var html string

		switch lib {
		case UILibRedoc:
			html = fmt.Sprintf(redocTemplate, docPath)
		case UILibScalar:
			html = fmt.Sprintf(scalarTemplate, docPath)
		default:
			panic(fmt.Sprintf("unsupported UI library: %s", lib))
		}

		ctx.SetResponseHeader("Content-Type", "text/html; charset=utf-8")
		ctx.SetResponseStatus(http.StatusOK)
		_, _ = ctx.ResponseBodyWriter().Write([]byte(html))
	}
}
