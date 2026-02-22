package echov4

import (
	"context"
	"crypto/tls"
	"io"
	"iter"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/cirius-go/miniapi"
)

// ContextImpl is the implementation of Context for Echo v4.
type ContextImpl struct {
	status int

	MiniAPIRoute miniapi.Route
	Origin       echo.Context
}

// RequestTLS implements miniapi.Context.
func (c *ContextImpl) RequestTLS() *tls.ConnectionState {
	return c.Origin.Request().TLS
}

// ResponseBodyWriter implements miniapi.Context.
func (c *ContextImpl) ResponseBodyWriter() io.Writer {
	return c.Origin.Response()
}

// AppendResponseHeader implements miniapi.Context.
func (c *ContextImpl) AppendResponseHeader(name string, value string) {
	c.Origin.Response().Header().Add(name, value)
}

// SetResponseHeader implements miniapi.Context.
func (c *ContextImpl) SetResponseHeader(name string, value string) {
	c.Origin.Response().Header().Set(name, value)
}

// ResponseStatus implements miniapi.Context.
func (c *ContextImpl) ResponseStatus() int {
	return c.status
}

// SetResponseStatus implements miniapi.Context.
func (c *ContextImpl) SetResponseStatus(status int) {
	c.status = status
	c.Origin.Response().WriteHeader(status)
}

// RequestMultipartForm implements miniapi.Context.
func (c *ContextImpl) RequestMultipartForm(maxMemory int64) (*multipart.Form, error) {
	err := c.Origin.Request().ParseMultipartForm(maxMemory)
	return c.Origin.Request().MultipartForm, err
}

// RequestBodyReader implements miniapi.Context.
func (c *ContextImpl) RequestBodyReader() io.ReadCloser {
	return c.Origin.Request().Body
}

// IterRequestHeader implements miniapi.Context.
func (c *ContextImpl) IterRequestHeader() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for name, values := range c.Origin.Request().Header {
			for _, value := range values {
				if ok := yield(name, value); !ok {
					return
				}
			}
		}
	}
}

// RequestHeader implements miniapi.Context.
func (c *ContextImpl) RequestHeader(name string) string {
	return c.Origin.Request().Header.Get(name)
}

// RequestQuery implements miniapi.Context.
func (c *ContextImpl) RequestQuery(name string) string {
	return c.Origin.QueryParam(name)
}

// RequestParam implements miniapi.Context.
func (c *ContextImpl) RequestParam(name string) string {
	return c.Origin.Param(name)
}

// RequestURL implements miniapi.Context.
func (c *ContextImpl) RequestURL() url.URL {
	return *c.Origin.Request().URL
}

// RemoteAddress implements miniapi.Context.
func (c *ContextImpl) RemoteAddress() string {
	return c.Origin.Request().RemoteAddr
}

// RequestHost implements miniapi.Context.
func (c *ContextImpl) RequestHost() string {
	return c.Origin.Request().Host
}

// RequestMethod implements miniapi.Context.
func (c *ContextImpl) RequestMethod() string {
	return c.Origin.Request().Method
}

// RequestContext implements miniapi.RequestContext.
func (c *ContextImpl) RequestContext() context.Context {
	return c.Origin.Request().Context()
}

// Route implements miniapi.Context.
func (c *ContextImpl) Route() miniapi.Route {
	return c.MiniAPIRoute
}

var _ miniapi.Context = (*ContextImpl)(nil)

// Router is the router for Echo v4.
type Router interface {
	Add(method, path string, handler echo.HandlerFunc, middleware ...echo.MiddlewareFunc) *echo.Route
}

// Adapter is the adapter for Echo v4.
type Adapter struct {
	httpHandler http.Handler
	router      Router
}

// AddRoute adds a route to the Echo v4 router.
func (a *Adapter) AddRoute(route miniapi.Route, handler miniapi.HandlerFunc) {
	path := route.Path()
	path = strings.ReplaceAll(path, "{", ":")
	path = strings.ReplaceAll(path, "}", "")
	a.router.Add(route.Method(), path, func(c echo.Context) error {
		ctx := &ContextImpl{MiniAPIRoute: route, Origin: c}
		handler(ctx)
		return nil
	})
}

// NewAdapter creates a new Adapter for Echo v4.
func NewAdapter(router *echo.Echo) *Adapter {
	return &Adapter{
		httpHandler: router,
		router:      router,
	}
}
