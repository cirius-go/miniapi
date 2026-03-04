package route

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"net/url"
	"testing"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/mocks"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

type mockContext struct {
	miniapi.Context
	requestBody  io.ReadCloser
	responseBody *bytes.Buffer
	status       int
	headers      map[string]string
	route        miniapi.Route
}

func (m *mockContext) Route() miniapi.Route             { return m.route }
func (m *mockContext) RequestContext() context.Context  { return context.Background() }
func (m *mockContext) RequestBodyReader() io.ReadCloser { return m.requestBody }
func (m *mockContext) ResponseBodyWriter() io.Writer    { return m.responseBody }
func (m *mockContext) SetResponseStatus(status int)     { m.status = status }
func (m *mockContext) ResponseStatus() int              { return m.status }
func (m *mockContext) SetResponseHeader(name, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[name] = value
}
func (m *mockContext) ResponseHeader(name string) string { return m.headers[name] }
func (m *mockContext) RequestHeader(name string) string  { return m.headers[name] }
func (m *mockContext) RequestURL() url.URL               { return url.URL{} }
func (m *mockContext) RequestParam(name string) string   { return "" }
func (m *mockContext) IterRequestHeader() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for k, v := range m.headers {
			if !yield(k, v) {
				return
			}
		}
	}
}

type testReq struct {
	Body struct {
		Name string `json:"name"`
	}
}

type testRes struct {
	Message string `json:"message"`
}

func testHandler(ctx context.Context, req *testReq) (*testRes, error) {
	if req.Body.Name == "error" {
		return nil, fmt.Errorf("triggered error")
	}
	return &testRes{Message: "Hello " + req.Body.Name}, nil
}

func TestTypedHandlerConversion(t *testing.T) {
	spec := Spec{Path: "/test", Method: "POST"}

	convey.Convey("Given a typed route", t, func() {
		mockBinder := mocks.NewBinder(t)
		r := New(spec, testHandler)
		r.SetBinder(mockBinder) // Inject mock

		resBuf := &bytes.Buffer{}
		mctx := &mockContext{
			responseBody: resBuf,
			headers:      make(map[string]string),
			route:        r,
		}

		convey.Convey("When a valid JSON request is received (T002)", func() {
			mockBinder.On("BindRequest", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				req := args.Get(1).(*testReq)
				req.Body.Name = "World"
			}).Return(nil)

			mockBinder.On("MarshalResponse", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				res := args.Get(1).(*testRes)
				mctx.SetResponseHeader("Content-Type", "application/json")
				_, _ = mctx.ResponseBodyWriter().Write([]byte(fmt.Sprintf(`{"message":"%s"}`, res.Message)))
			}).Return(nil)
			r.HandlerFunc()(mctx)

			convey.Convey("Then it should respond with 200 OK and correct JSON body", func() {
				convey.So(mctx.status, convey.ShouldEqual, 200)
				convey.So(mctx.headers["Content-Type"], convey.ShouldEqual, "application/json")
				convey.So(resBuf.String(), convey.ShouldEqual, `{"message":"Hello World"}`)
				mockBinder.AssertExpectations(t)
			})
		})

		convey.Convey("When an empty request body is received (T003)", func() {
			mockBinder.On("BindRequest", mock.Anything, mock.Anything).Return(nil)
			mockBinder.On("MarshalResponse", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				res := args.Get(1).(*testRes)
				mctx.SetResponseHeader("Content-Type", "application/json")
				_, _ = mctx.ResponseBodyWriter().Write([]byte(fmt.Sprintf(`{"message":"%s"}`, res.Message)))
			}).Return(nil)

			r.HandlerFunc()(mctx)

			convey.Convey("Then it should respond with 200 OK and default struct values", func() {
				convey.So(mctx.status, convey.ShouldEqual, 200)
				convey.So(resBuf.String(), convey.ShouldEqual, `{"message":"Hello "}`)
				mockBinder.AssertExpectations(t)
			})
		})

		convey.Convey("When an invalid JSON request is received (T004)", func() {
			mockBinder.On("BindRequest", mock.Anything, mock.Anything).Return(fmt.Errorf("binding failed: invalid json"))

			r.HandlerFunc()(mctx)

			convey.Convey("Then it should respond with 400 Bad Request", func() {
				convey.So(mctx.status, convey.ShouldEqual, 400)
				convey.So(mctx.headers["Content-Type"], convey.ShouldEqual, "application/problem+json")
				// MarshalResponse should NOT be called on a binding error.
				mockBinder.AssertExpectations(t)
			})
		})

		convey.Convey("When the handler returns an error (T005)", func() {
			mockBinder.On("BindRequest", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				req := args.Get(1).(*testReq)
				req.Body.Name = "error"
			}).Return(nil)

			r.HandlerFunc()(mctx)

			convey.Convey("Then it should respond with 500 Internal Server Error and application/problem+json", func() {
				convey.So(mctx.status, convey.ShouldEqual, 500)
				convey.So(mctx.headers["Content-Type"], convey.ShouldEqual, "application/problem+json")
				// MarshalResponse should NOT be called when the handler errors.
				mockBinder.AssertExpectations(t)
			})
		})

		convey.Convey("When a custom Content-Type is already set (T006)", func() {
			mockBinder.On("BindRequest", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				req := args.Get(1).(*testReq)
				req.Body.Name = "World"
			}).Return(nil)

			mockBinder.On("MarshalResponse", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				if mctx.ResponseHeader("Content-Type") == "" {
					mctx.SetResponseHeader("Content-Type", "application/json")
				}
			}).Return(nil)

			mctx.SetResponseHeader("Content-Type", "text/plain")
			r.HandlerFunc()(mctx)

			convey.Convey("Then it should not be overwritten by application/json", func() {
				convey.So(mctx.headers["Content-Type"], convey.ShouldEqual, "text/plain")
				mockBinder.AssertExpectations(t)
			})
		})
	})
}
