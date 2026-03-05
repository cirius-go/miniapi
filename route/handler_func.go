package route

import (
	"encoding/json"
	"net/http"

	"github.com/cirius-go/miniapi"
)

// DefaultErrorEncoder is the default error encoder for miniapi. It encodes the
// error as a JSON object with an "error" field.
func DefaultErrorEncoder(c miniapi.Context, err error) {
	c.SetResponseHeader("Content-Type", "application/problem+json")

	if c.ResponseStatus() == 0 {
		c.SetResponseStatus(http.StatusInternalServerError)
	}

	_ = json.NewEncoder(c.ResponseBodyWriter()).Encode(map[string]any{
		"error": err.Error(),
	})
}

// MakeHandlerFuncBuilder creates a HandlerFuncBuilder from a typed handler
// function, a Binder, and an ErrorEncoder.
func MakeHandlerFuncBuilder[Rq, Rp any](fn miniapi.TypedHandlerFunc[Rq, Rp]) miniapi.HandlerFuncBuilder {
	return func(b miniapi.Binder, e miniapi.ErrorEncoder) miniapi.HandlerFunc {
		return func(c miniapi.Context) {
			var (
				ctx = c.RequestContext()
				req = new(Rq)
			)

			// Decode request
			if err := b.BindRequest(c, req); err != nil {
				e(c, err)
				return
			}

			// Call handler
			res, err := fn(ctx, req)
			if err != nil {
				e(c, err)
				return
			}

			// Success
			if c.ResponseStatus() == 0 {
				c.SetResponseStatus(http.StatusOK)
			}
			_ = b.MarshalResponse(c, res)
		}
	}
}
