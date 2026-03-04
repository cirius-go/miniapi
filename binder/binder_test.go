package binder

import (
	"bytes"
	"io"
	"iter"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/cirius-go/miniapi"
	"github.com/smartystreets/goconvey/convey"
)

func TestTypeConversions(t *testing.T) {
	convey.Convey("Given setPrimitiveValue", t, func() {
		convey.Convey("It should correctly parse strings", func() {
			var s string
			err := setPrimitiveValue(reflect.ValueOf(&s), "hello")
			convey.So(err, convey.ShouldBeNil)
			convey.So(s, convey.ShouldEqual, "hello")
		})

		convey.Convey("It should correctly parse integers", func() {
			var i int
			err := setPrimitiveValue(reflect.ValueOf(&i), "42")
			convey.So(err, convey.ShouldBeNil)
			convey.So(i, convey.ShouldEqual, 42)
		})

		convey.Convey("It should correctly parse booleans", func() {
			var b bool
			err := setPrimitiveValue(reflect.ValueOf(&b), "true")
			convey.So(err, convey.ShouldBeNil)
			convey.So(b, convey.ShouldBeTrue)
		})

		convey.Convey("It should return ErrTypeMismatch for invalid booleans", func() {
			var b bool
			err := setPrimitiveValue(reflect.ValueOf(&b), "notabool")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "type mismatch")
		})
	})

	convey.Convey("Given setSliceValue", t, func() {
		convey.Convey("It should correctly parse a slice of strings", func() {
			var s []string
			err := setSliceValue(reflect.ValueOf(&s), []string{"a", "b"})
			convey.So(err, convey.ShouldBeNil)
			convey.So(s, convey.ShouldResemble, []string{"a", "b"})
		})

		convey.Convey("It should correctly parse a slice of ints", func() {
			var s []int
			err := setSliceValue(reflect.ValueOf(&s), []string{"1", "2"})
			convey.So(err, convey.ShouldBeNil)
			convey.So(s, convey.ShouldResemble, []int{1, 2})
		})
	})

	convey.Convey("Given extractStructTags", t, func() {
		convey.Convey("It should correctly identify tags and Body field", func() {
			type TestStruct struct {
				ID    string   `path:"id"`
				Sort  []string `query:"sort"`
				Token string   `header:"X-Auth"`
				Body  struct {
					Name string
				}
				Ignored int
			}

			pathMeta, queryMeta, headerMeta, bodyIdx := extractStructTags(reflect.TypeOf(TestStruct{}))
			convey.So(len(pathMeta), convey.ShouldEqual, 1)
			convey.So(pathMeta[0].Tag, convey.ShouldEqual, "id")

			convey.So(len(queryMeta), convey.ShouldEqual, 1)
			convey.So(queryMeta[0].Tag, convey.ShouldEqual, "sort")
			convey.So(queryMeta[0].IsSlice, convey.ShouldBeTrue)

			convey.So(len(headerMeta), convey.ShouldEqual, 1)
			convey.So(headerMeta[0].Tag, convey.ShouldEqual, "X-Auth")

			convey.So(len(bodyIdx), convey.ShouldEqual, 1)
		})
	})
}

// binderMockContext implements miniapi.Context for testing purposes
type binderMockContext struct {
	miniapi.Context // Embed interface to panic on unimplemented methods
	pathParams      map[string]string
	queryParams     map[string][]string
	headers         map[string][]string
	bodyReader      io.ReadCloser
	responseHeaders map[string]string
	responseWriter  *bytes.Buffer
}

func (m *binderMockContext) ResponseHeader(name string) string {
	if m.responseHeaders == nil {
		return ""
	}
	return m.responseHeaders[name]
}

func (m *binderMockContext) SetResponseHeader(name, value string) {
	if m.responseHeaders == nil {
		m.responseHeaders = make(map[string]string)
	}
	m.responseHeaders[name] = value
}

func (m *binderMockContext) ResponseBodyWriter() io.Writer {
	if m.responseWriter == nil {
		m.responseWriter = new(bytes.Buffer)
	}
	return m.responseWriter
}

func (m *binderMockContext) RequestParam(name string) string {
	return m.pathParams[name]
}

func (m *binderMockContext) RequestHeader(name string) string {
	if vals, ok := m.headers[name]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (m *binderMockContext) RequestURL() url.URL {
	u := url.URL{}
	q := u.Query()
	for k, vals := range m.queryParams {
		for _, v := range vals {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u
}

func (m *binderMockContext) IterRequestHeader() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for k, vals := range m.headers {
			for _, v := range vals {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

func (m *binderMockContext) RequestBodyReader() io.ReadCloser {
	return m.bodyReader
}

func TestBindRequest(t *testing.T) {
	convey.Convey("Given BindRequest", t, func() {
		convey.Convey("It should bind path, query, header, and JSON body correctly", func() {
			type TestReq struct {
				ID   string   `path:"id"`
				Age  int      `query:"age"`
				Tags []string `query:"tags"`
				Auth string   `header:"Authorization"`
				Body struct {
					Name string `json:"name"`
				}
			}

			body := strings.NewReader(`{"name":"test user"}`)
			ctx := &binderMockContext{
				pathParams:  map[string]string{"id": "123"},
				queryParams: map[string][]string{"age": {"25"}, "tags": {"go", "api"}},
				headers:     map[string][]string{"Authorization": {"Bearer token"}},
				bodyReader:  io.NopCloser(body),
			}

			var req TestReq
			binder := New(DefaultBindingOptions)
			err := binder.BindRequest(ctx, &req)
			convey.So(err, convey.ShouldBeNil)

			convey.So(req.ID, convey.ShouldEqual, "123")
			convey.So(req.Age, convey.ShouldEqual, 25)
			convey.So(req.Tags, convey.ShouldResemble, []string{"go", "api"})
			convey.So(req.Auth, convey.ShouldEqual, "Bearer token")
			convey.So(req.Body.Name, convey.ShouldEqual, "test user")
		})
	})
}

func TestMarshalResponse(t *testing.T) {
	convey.Convey("Given marshalResponse", t, func() {
		convey.Convey("It should correctly marshal to JSON and set Content-Type", func() {
			ctx := &binderMockContext{}
			res := struct {
				Message string `json:"message"`
			}{Message: "success"}

			binder := New(DefaultBindingOptions)
			err := binder.MarshalResponse(ctx, &res)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ctx.ResponseHeader("Content-Type"), convey.ShouldEqual, "application/json")
			convey.So(ctx.responseWriter.String(), convey.ShouldContainSubstring, `{"message":"success"}`)
		})

		convey.Convey("It should not overwrite an existing Content-Type", func() {
			ctx := &binderMockContext{}
			ctx.SetResponseHeader("Content-Type", "application/custom+json")
			res := struct {
				Message string `json:"message"`
			}{Message: "success"}

			binder := New(DefaultBindingOptions)
			err := binder.MarshalResponse(ctx, &res)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ctx.ResponseHeader("Content-Type"), convey.ShouldEqual, "application/custom+json")
		})

		convey.Convey("It should return nil if response is nil", func() {
			ctx := &binderMockContext{}
			binder := New(DefaultBindingOptions)
			err := binder.MarshalResponse(ctx, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ctx.responseWriter, convey.ShouldBeNil)
		})
	})
}

func TestContentNegotiation(t *testing.T) {
	convey.Convey("Given BindRequest with Content-Type", t, func() {
		convey.Convey("It should reject unsupported Content-Type", func() {
			type TestReq struct {
				Body struct{ Name string }
			}
			ctx := &binderMockContext{
				headers:    map[string][]string{"Content-Type": {"application/xml"}},
				bodyReader: io.NopCloser(strings.NewReader(`<name>test</name>`)),
			}
			var req TestReq
			binder := New(DefaultBindingOptions)
			err := binder.BindRequest(ctx, &req)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "unsupported media type")
		})
	})

	convey.Convey("Given marshalResponse with Accept Header", t, func() {
		convey.Convey("It should output Plain Text if text/plain is requested", func() {
			ctx := &binderMockContext{
				headers: map[string][]string{"Accept": {"text/plain"}},
			}
			res := "hello plain text"
			binder := New(DefaultBindingOptions)
			err := binder.MarshalResponse(ctx, &res)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ctx.ResponseHeader("Content-Type"), convey.ShouldEqual, "text/plain")
			convey.So(ctx.responseWriter.String(), convey.ShouldEqual, "hello plain text")
		})

		convey.Convey("It should default to JSON if no Accept header is present", func() {
			ctx := &binderMockContext{
				headers: map[string][]string{},
			}
			res := struct{ Message string }{Message: "test"}
			binder := New(DefaultBindingOptions)
			err := binder.MarshalResponse(ctx, &res)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func BenchmarkBindRequest(b *testing.B) {
	type TestReq struct {
		ID   string   `path:"id"`
		Age  int      `query:"age"`
		Tags []string `query:"tags"`
		Auth string   `header:"Authorization"`
		Body struct {
			Name string `json:"name"`
		}
	}

	bodyData := `{"name":"benchmark test"}`

	b.ReportAllocs()

	for b.Loop() {
		// Re-create context for each run to simulate real requests and avoid EOF on body
		ctx := &binderMockContext{
			pathParams:  map[string]string{"id": "123"},
			queryParams: map[string][]string{"age": {"25"}, "tags": {"go", "api"}},
			headers:     map[string][]string{"Authorization": {"Bearer token"}, "Content-Type": {"application/json"}},
			bodyReader:  io.NopCloser(strings.NewReader(bodyData)),
		}

		var req TestReq
		binder := New(DefaultBindingOptions)
		_ = binder.BindRequest(ctx, &req)
	}
}
