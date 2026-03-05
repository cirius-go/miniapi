package binder

import (
	"bytes"
	"errors"
	"io"
	"iter"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/cirius-go/miniapi/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

// Test Fixtures

type PrimitiveStruct struct {
	String string  `path:"str"`
	Int    int     `query:"int"`
	Float  float64 `header:"X-Float"`
	Bool   bool    `query:"bool"`
}

type SliceStruct struct {
	Strings []string `query:"strs"`
	Ints    []int    `header:"X-Ints"`
}

type PointerStruct struct {
	String *string `path:"str"`
	Int    *int    `query:"int"`
}

type JSONBodyStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type BodyStruct struct {
	Body JSONBodyStruct
}

type RawBodyReaderStruct struct {
	Body io.Reader
}

type RawBodyByteStruct struct {
	Body []byte
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (e *errorReader) Close() error {
	return nil
}

type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func setupMockContext(ctx *mocks.Context) {
	u, _ := url.Parse("http://localhost")
	ctx.On("RequestURL").Return(*u).Maybe()
	ctx.On("RequestParam", mock.Anything).Return("").Maybe()
	ctx.On("IterRequestHeader").Return(iter.Seq2[string, string](func(yield func(string, string) bool) {})).Maybe()
}

func TestBinder(t *testing.T) {
	Convey("Given a Binder instance", t, func() {
		b := New(DefaultConfig())
		ctx := new(mocks.Context)

		Convey("When mapping path parameters to primitives", func() {
			req := &PrimitiveStruct{}
			ctx.On("RequestParam", "str").Return("hello")

			err := b.BindPathParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "String", Tag: "str", FieldIndex: []int{0}},
			})

			So(err, ShouldBeNil)
			So(req.String, ShouldEqual, "hello")
		})

		Convey("When mapping query parameters", func() {
			req := &PrimitiveStruct{}
			u, _ := url.Parse("http://localhost?int=42&bool=true")
			ctx.On("RequestURL").Return(*u)

			err := b.BindQueryParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Int", Tag: "int", FieldIndex: []int{1}},
				{Name: "Bool", Tag: "bool", FieldIndex: []int{3}},
			})

			So(err, ShouldBeNil)
			So(req.Int, ShouldEqual, 42)
			So(req.Bool, ShouldBeTrue)
		})

		Convey("When mapping query parameter slices", func() {
			req := &SliceStruct{}
			u, _ := url.Parse("http://localhost?strs=a&strs=b")
			ctx.On("RequestURL").Return(*u)

			err := b.BindQueryParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Strings", Tag: "strs", FieldIndex: []int{0}, IsSlice: true},
			})

			So(err, ShouldBeNil)
			So(req.Strings, ShouldResemble, []string{"a", "b"})
		})

		Convey("When mapping header parameters", func() {
			req := &PrimitiveStruct{}
			ctx.On("IterRequestHeader").Return(iter.Seq2[string, string](func(yield func(string, string) bool) {
				if !yield("X-Float", "3.14") {
					return
				}
			}))

			err := b.BindHeaderParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Float", Tag: "X-Float", FieldIndex: []int{2}},
			})

			So(err, ShouldBeNil)
			So(req.Float, ShouldEqual, 3.14)
		})

		Convey("When mapping header parameter slices", func() {
			req := &SliceStruct{}
			ctx.On("IterRequestHeader").Return(iter.Seq2[string, string](func(yield func(string, string) bool) {
				if !yield("X-Ints", "1") {
					return
				}
				if !yield("X-Ints", "2") {
					return
				}
			}))

			err := b.BindHeaderParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Ints", Tag: "X-Ints", FieldIndex: []int{1}, IsSlice: true},
			})

			So(err, ShouldBeNil)
			So(req.Ints, ShouldResemble, []int{1, 2})
		})

		Convey("When mapping JSON body", func() {
			req := &BodyStruct{}
			bodyJSON := `{"name":"John","age":30}`
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(bodyJSON)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json")

			err := b.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})

			So(err, ShouldBeNil)
			So(req.Body.Name, ShouldEqual, "John")
			So(req.Body.Age, ShouldEqual, 30)
		})

		Convey("When mapping to pointer fields", func() {
			req := &PointerStruct{}
			ctx.On("RequestParam", "str").Return("ptr_hello")
			u, _ := url.Parse("http://localhost?int=100")
			ctx.On("RequestURL").Return(*u)

			err := b.BindPathParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "String", Tag: "str", FieldIndex: []int{0}},
			})
			So(err, ShouldBeNil)
			So(*req.String, ShouldEqual, "ptr_hello")

			err = b.BindQueryParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Int", Tag: "int", FieldIndex: []int{1}},
			})
			So(err, ShouldBeNil)
			So(*req.Int, ShouldEqual, 100)
		})

		Convey("When mapping request body to io.Reader", func() {
			req := &RawBodyReaderStruct{}
			bodyContent := "raw data"
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(bodyContent)))

			err := b.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})

			So(err, ShouldBeNil)
			data, _ := io.ReadAll(req.Body)
			So(string(data), ShouldEqual, bodyContent)
		})

		Convey("When mapping request body to []byte", func() {
			req := &RawBodyByteStruct{}
			bodyContent := "byte data"
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(bodyContent)))

			err := b.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})

			So(err, ShouldBeNil)
			So(string(req.Body), ShouldEqual, bodyContent)
		})

		Convey("When mapping missing values", func() {
			req := &PrimitiveStruct{String: "original", Int: 42}
			ctx.On("RequestParam", "str").Return("")
			u, _ := url.Parse("http://localhost")
			ctx.On("RequestURL").Return(*u)

			err := b.BindPathParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "String", Tag: "str", FieldIndex: []int{0}},
			})
			So(err, ShouldBeNil)
			So(req.String, ShouldEqual, "original")

			err = b.BindQueryParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Int", Tag: "int", FieldIndex: []int{1}},
			})
			So(err, ShouldBeNil)
			So(req.Int, ShouldEqual, 42)
		})

		Convey("When mapping invalid primitive types", func() {
			req := &PrimitiveStruct{}
			u, _ := url.Parse("http://localhost?int=abc")
			ctx.On("RequestURL").Return(*u)

			err := b.BindQueryParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Int", Tag: "int", FieldIndex: []int{1}},
			})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTypeMismatch), ShouldBeTrue)
		})

		Convey("When body exceeds MaxBodySize", func() {
			cfg := DefaultConfig()
			cfg.MaxBodySize = 5
			b2 := New(cfg)

			req := &RawBodyByteStruct{}
			bodyContent := "too long body"
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(bodyContent)))

			err := b2.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrPayloadTooLarge), ShouldBeTrue)
		})

		Convey("When JSON body is malformed", func() {
			req := &BodyStruct{}
			bodyJSON := `{"name":"John", "age": "not-an-int"}`
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(bodyJSON)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json")

			err := b.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrBindingFailed), ShouldBeTrue)
		})

		Convey("When Content-Type is unsupported", func() {
			req := &BodyStruct{}
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(`{}`)))
			ctx.On("RequestHeader", "Content-Type").Return("application/xml")

			err := b.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrUnsupportedMediaType), ShouldBeTrue)
		})

		Convey("When DisallowUnknownFields is enabled", func() {
			cfg := DefaultConfig()
			cfg.DisallowUnknownFields = true
			b2 := New(cfg)

			req := &BodyStruct{}
			bodyJSON := `{"name":"John", "age": 30, "unknown": "field"}`
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(bodyJSON)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json")

			err := b2.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrBindingFailed), ShouldBeTrue)
		})

		Convey("When req is not a pointer to a struct", func() {
			err := b.BindRequest(ctx, PrimitiveStruct{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrBindingFailed), ShouldBeTrue)

			str := "not a struct"
			err = b.BindRequest(ctx, &str)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrBindingFailed), ShouldBeTrue)
		})

		Convey("When marshaling response to JSON", func() {
			res := &JSONBodyStruct{Name: "Alice", Age: 25}
			ctx.On("RequestHeader", "Accept").Return("application/json")
			ctx.On("ResponseHeader", "Content-Type").Return("")
			ctx.On("SetResponseHeader", "Content-Type", "application/json")

			var buf bytes.Buffer
			ctx.On("ResponseBodyWriter").Return(&buf)

			err := b.MarshalResponse(ctx, res)
			So(err, ShouldBeNil)
			So(buf.String(), ShouldEqual, `{"name":"Alice","age":25}`+"\n")
		})

		Convey("When marshaling response to text/plain", func() {
			ctx.On("RequestHeader", "Accept").Return("text/plain")
			ctx.On("ResponseHeader", "Content-Type").Return("")
			ctx.On("SetResponseHeader", "Content-Type", "text/plain")

			Convey("Given a string", func() {
				res := "plain text"
				var buf bytes.Buffer
				ctx.On("ResponseBodyWriter").Return(&buf)

				err := b.MarshalResponse(ctx, res)
				So(err, ShouldBeNil)
				So(buf.String(), ShouldEqual, "plain text")
			})

			Convey("Given a pointer to a string", func() {
				s := "ptr text"
				res := &s
				var buf bytes.Buffer
				ctx.On("ResponseBodyWriter").Return(&buf)

				err := b.MarshalResponse(ctx, res)
				So(err, ShouldBeNil)
				So(buf.String(), ShouldEqual, "ptr text")
			})

			Convey("Given other type", func() {
				res := 123
				var buf bytes.Buffer
				ctx.On("ResponseBodyWriter").Return(&buf)

				err := b.MarshalResponse(ctx, res)
				So(err, ShouldBeNil)
				So(buf.String(), ShouldEqual, "123")
			})
		})

		Convey("When no Accept header is present", func() {
			res := &JSONBodyStruct{Name: "Default", Age: 0}
			ctx.On("RequestHeader", "Accept").Return("")
			ctx.On("ResponseHeader", "Content-Type").Return("")
			ctx.On("SetResponseHeader", "Content-Type", "application/json")

			var buf bytes.Buffer
			ctx.On("ResponseBodyWriter").Return(&buf)

			err := b.MarshalResponse(ctx, res)
			So(err, ShouldBeNil)
			So(buf.String(), ShouldEqual, `{"name":"Default","age":0}`+"\n")
		})

		Convey("When response is nil", func() {
			err := b.MarshalResponse(ctx, nil)
			So(err, ShouldBeNil)
		})

		Convey("When using BindRequest for full struct binding", func() {
			type FullStruct struct {
				Path   string  `path:"p"`
				Query  int     `query:"q"`
				Header float64 `header:"h"`
				Body   struct {
					Foo string `json:"foo"`
				}
			}
			req := &FullStruct{}

			ctx.On("RequestParam", "p").Return("p_val")

			u, _ := url.Parse("http://localhost?q=123")
			ctx.On("RequestURL").Return(*u)

			ctx.On("IterRequestHeader").Return(iter.Seq2[string, string](func(yield func(string, string) bool) {
				_ = yield("h", "1.23")
			}))

			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(`{"foo":"bar"}`)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json")

			err := b.BindRequest(ctx, req)
			So(err, ShouldBeNil)
			So(req.Path, ShouldEqual, "p_val")
			So(req.Query, ShouldEqual, 123)
			So(req.Header, ShouldEqual, 1.23)
			So(req.Body.Foo, ShouldEqual, "bar")
		})

		Convey("When mapping all primitive types", func() {
			type AllPrimitives struct {
				Int8    int8    `query:"i8"`
				Int16   int16   `query:"i16"`
				Int32   int32   `query:"i32"`
				Int64   int64   `query:"i64"`
				Uint    uint    `query:"u"`
				Uint8   uint8   `query:"u8"`
				Uint16  uint16  `query:"u16"`
				Uint32  uint32  `query:"u32"`
				Uint64  uint64  `query:"u64"`
				Float32 float32 `query:"f32"`
			}
			req := &AllPrimitives{}
			u, _ := url.Parse("http://localhost?i8=8&i16=16&i32=32&i64=64&u=1&u8=8&u16=16&u32=32&u64=64&f32=0.5")
			ctx.On("RequestURL").Return(*u)

			err := b.BindRequest(ctx, req)
			So(err, ShouldBeNil)
			So(req.Int8, ShouldEqual, 8)
			So(req.Int16, ShouldEqual, 16)
			So(req.Int32, ShouldEqual, 32)
			So(req.Int64, ShouldEqual, 64)
			So(req.Uint, ShouldEqual, 1)
			So(req.Uint8, ShouldEqual, 8)
			So(req.Uint16, ShouldEqual, 16)
			So(req.Uint32, ShouldEqual, 32)
			So(req.Uint64, ShouldEqual, 64)
			So(req.Float32, ShouldEqual, 0.5)
		})

		Convey("When BindRequest fails in path params", func() {
			// To trigger error in SetPrimitiveValue we need invalid type
			// But path parameters are usually strings.
			// Wait, let's use Int in path if possible.
			type IntPath struct {
				Int int `path:"i"`
			}
			req2 := &IntPath{}
			ctx.On("RequestParam", "i").Return("abc")
			err := b.BindRequest(ctx, req2)
			So(err, ShouldNotBeNil)
		})

		Convey("When BindRequest fails in query params", func() {
			setupMockContext(ctx)
			u, _ := url.Parse("http://localhost?int=abc")
			ctx.On("RequestURL").Return(*u).Unset() // Override the one in setupMockContext
			ctx.On("RequestURL").Return(*u)
			err := b.BindRequest(ctx, &PrimitiveStruct{})
			So(err, ShouldNotBeNil)
		})

		Convey("When BindRequest fails in header params", func() {
			setupMockContext(ctx)
			ctx.On("IterRequestHeader").Return(iter.Seq2[string, string](func(yield func(string, string) bool) {
				_ = yield("X-Float", "abc")
			})).Unset()
			ctx.On("IterRequestHeader").Return(iter.Seq2[string, string](func(yield func(string, string) bool) {
				_ = yield("X-Float", "abc")
			}))
			err := b.BindRequest(ctx, &PrimitiveStruct{})
			So(err, ShouldNotBeNil)
		})

		Convey("When content negotiation prefers plain text but JSON is also present", func() {
			ctx.On("RequestHeader", "Accept").Return("text/plain, application/json;q=0.5")
			ctx.On("ResponseHeader", "Content-Type").Return("")
			ctx.On("SetResponseHeader", "Content-Type", "text/plain")
			var buf bytes.Buffer
			ctx.On("ResponseBodyWriter").Return(&buf)
			err := b.MarshalResponse(ctx, "hello")
			So(err, ShouldBeNil)
			So(buf.String(), ShouldEqual, "hello")
		})

		Convey("When content negotiation prefers JSON over plain text", func() {
			ctx.On("RequestHeader", "Accept").Return("application/json, text/plain;q=0.5")
			ctx.On("ResponseHeader", "Content-Type").Return("")
			ctx.On("SetResponseHeader", "Content-Type", "application/json")
			var buf bytes.Buffer
			ctx.On("ResponseBodyWriter").Return(&buf)
			err := b.MarshalResponse(ctx, map[string]string{"foo": "bar"})
			So(err, ShouldBeNil)
			So(buf.String(), ShouldEqual, `{"foo":"bar"}`+"\n")
		})

		Convey("When Content-Type is already set in MarshalResponse", func() {
			ctx.On("RequestHeader", "Accept").Return("application/json")
			ctx.On("ResponseHeader", "Content-Type").Return("application/json")
			var buf bytes.Buffer
			ctx.On("ResponseBodyWriter").Return(&buf)
			err := b.MarshalResponse(ctx, map[string]string{"foo": "bar"})
			So(err, ShouldBeNil)
		})

		Convey("When RequestBodyReader is nil", func() {
			ctx.On("RequestBodyReader").Return(nil)
			err := b.BindBody(ctx, reflect.ValueOf(&RawBodyByteStruct{}).Elem(), []int{0})
			So(err, ShouldBeNil)
		})

		Convey("When io.ReadAll fails in BindBody for io.Reader", func() {
			ctx.On("RequestBodyReader").Return(&errorReader{})
			err := b.BindBody(ctx, reflect.ValueOf(&RawBodyReaderStruct{}).Elem(), []int{0})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrBindingFailed), ShouldBeTrue)
		})

		Convey("When io.ReadAll fails in BindBody for []byte", func() {
			ctx.On("RequestBodyReader").Return(&errorReader{})
			err := b.BindBody(ctx, reflect.ValueOf(&RawBodyByteStruct{}).Elem(), []int{0})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrBindingFailed), ShouldBeTrue)
		})

		Convey("When payload-too-large fails in BindBody for io.Reader", func() {
			cfg := DefaultConfig()
			cfg.MaxBodySize = 5
			b2 := New(cfg)
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader("too long")))
			err := b2.BindBody(ctx, reflect.ValueOf(&RawBodyReaderStruct{}).Elem(), []int{0})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrPayloadTooLarge), ShouldBeTrue)
		})

		Convey("When Accept header doesn't contain application/json", func() {
			ctx.On("RequestHeader", "Accept").Return("text/plain")
			ctx.On("ResponseHeader", "Content-Type").Return("")
			ctx.On("SetResponseHeader", "Content-Type", "text/plain")
			var buf bytes.Buffer
			ctx.On("ResponseBodyWriter").Return(&buf)
			err := b.MarshalResponse(ctx, "hello")
			So(err, ShouldBeNil)
			So(buf.String(), ShouldEqual, "hello")
		})

		Convey("When Body is a pointer and nil", func() {
			type PointerBody struct {
				Body *JSONBodyStruct
			}
			req := &PointerBody{}
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(`{"name":"test"}`)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json")
			err := b.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})
			So(err, ShouldBeNil)
			So(req.Body, ShouldNotBeNil)
			So(req.Body.Name, ShouldEqual, "test")
		})

		Convey("When ResponseBodyWriter fails in MarshalResponse", func() {
			ctx.On("RequestHeader", "Accept").Return("text/plain")
			ctx.On("ResponseHeader", "Content-Type").Return("text/plain")
			ctx.On("ResponseBodyWriter").Return(&errorWriter{})
			err := b.MarshalResponse(ctx, "hello")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrBindingFailed), ShouldBeTrue)
		})

		Convey("When JSON encoding fails in MarshalResponse", func() {
			ctx.On("RequestHeader", "Accept").Return("application/json")
			ctx.On("ResponseHeader", "Content-Type").Return("application/json")
			ctx.On("ResponseBodyWriter").Return(&errorWriter{})
			err := b.MarshalResponse(ctx, map[string]any{"foo": func() {}}) // Non-serializable
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrBindingFailed), ShouldBeTrue)
		})

		Convey("When header is missing", func() {
			req := &SliceStruct{}
			ctx.On("IterRequestHeader").Return(iter.Seq2[string, string](func(yield func(string, string) bool) {
				// No headers
			}))
			err := b.BindHeaderParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Ints", Tag: "X-Ints", FieldIndex: []int{1}, IsSlice: true},
			})
			So(err, ShouldBeNil)
			So(req.Ints, ShouldBeEmpty)
		})

		Convey("When header mapping fails for single value", func() {
			req := &PrimitiveStruct{}
			ctx.On("IterRequestHeader").Return(iter.Seq2[string, string](func(yield func(string, string) bool) {
				_ = yield("X-Float", "abc")
			}))
			err := b.BindHeaderParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Float", Tag: "X-Float", FieldIndex: []int{2}},
			})
			So(err, ShouldNotBeNil)
		})

		Convey("When query mapping fails for single value", func() {
			req := &PrimitiveStruct{}
			u, _ := url.Parse("http://localhost?int=abc")
			ctx.On("RequestURL").Return(*u)
			err := b.BindQueryParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Int", Tag: "int", FieldIndex: []int{1}},
			})
			So(err, ShouldNotBeNil)
		})

		Convey("When query mapping fails for slice", func() {
			// Let's use Ints slice in query.
			type IntsQuery struct {
				Ints []int `query:"i"`
			}
			req2 := &IntsQuery{}
			u2, _ := url.Parse("http://localhost?i=1&i=abc")
			ctx.On("RequestURL").Return(*u2)
			err := b.BindQueryParams(ctx, reflect.ValueOf(req2).Elem(), []StructFieldMeta{
				{Name: "Ints", Tag: "i", FieldIndex: []int{0}, IsSlice: true},
			})
			So(err, ShouldNotBeNil)
		})

		Convey("When header mapping fails for slice", func() {
			req := &SliceStruct{}
			ctx.On("IterRequestHeader").Return(iter.Seq2[string, string](func(yield func(string, string) bool) {
				_ = yield("X-Ints", "abc")
			}))
			err := b.BindHeaderParams(ctx, reflect.ValueOf(req).Elem(), []StructFieldMeta{
				{Name: "Ints", Tag: "X-Ints", FieldIndex: []int{1}, IsSlice: true},
			})
			So(err, ShouldNotBeNil)
		})

		Convey("When BindRequest fails in body binding", func() {
			setupMockContext(ctx)
			type BodyReq struct {
				Body struct{ Name string }
			}
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(`invalid json`)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json")
			err := b.BindRequest(ctx, &BodyReq{})
			So(err, ShouldNotBeNil)
		})

		Convey("When payload-too-large fails in BindBody for []byte", func() {
			cfg := DefaultConfig()
			cfg.MaxBodySize = 5
			b2 := New(cfg)
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader("too long")))
			err := b2.BindBody(ctx, reflect.ValueOf(&RawBodyByteStruct{}).Elem(), []int{0})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrPayloadTooLarge), ShouldBeTrue)
		})

		Convey("When payload-too-large fails in BindBody for JSON", func() {
			cfg := DefaultConfig()
			cfg.MaxBodySize = 5
			b2 := New(cfg)
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(`{"name":"too long"}`)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json")
			err := b2.BindBody(ctx, reflect.ValueOf(&BodyStruct{}).Elem(), []int{0})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrPayloadTooLarge), ShouldBeTrue)
		})

		Convey("When JSON body is empty (EOF)", func() {
			req := &BodyStruct{}
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(``)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json")
			err := b.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})
			So(err, ShouldBeNil)
		})

		Convey("When MaxBodySize is zero or negative", func() {
			cfg := DefaultConfig()
			cfg.MaxBodySize = 0
			b2 := New(cfg)
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(`{}`)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json")
			err := b2.BindBody(ctx, reflect.ValueOf(&BodyStruct{}).Elem(), []int{0})
			So(err, ShouldBeNil)
		})

		Convey("When Content-Type has charset", func() {
			req := &BodyStruct{}
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(`{"name":"test"}`)))
			ctx.On("RequestHeader", "Content-Type").Return("application/json; charset=utf-8")
			err := b.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})
			So(err, ShouldBeNil)
			So(req.Body.Name, ShouldEqual, "test")
		})

		Convey("When Content-Type is empty", func() {
			req := &BodyStruct{}
			ctx.On("RequestBodyReader").Return(io.NopCloser(strings.NewReader(`{"name":"test"}`)))
			ctx.On("RequestHeader", "Content-Type").Return("")
			err := b.BindBody(ctx, reflect.ValueOf(req).Elem(), []int{0})
			So(err, ShouldBeNil)
			So(req.Body.Name, ShouldEqual, "test")
		})
	})
}

func TestBinderInternal(t *testing.T) {
	Convey("Given internal binder helpers", t, func() {
		Convey("SetPrimitiveValue with unsupported types", func() {
			v := reflect.ValueOf(struct{}{})
			err := SetPrimitiveValue(v, "val")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTypeMismatch), ShouldBeTrue)
		})

		Convey("SetSliceValue with non-slice target", func() {
			var i int
			err := SetSliceValue(reflect.ValueOf(&i).Elem(), []string{"1"})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTypeMismatch), ShouldBeTrue)
		})

		Convey("SetSliceValue with primitive conversion error", func() {
			var is []int
			err := SetSliceValue(reflect.ValueOf(&is).Elem(), []string{"abc"})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTypeMismatch), ShouldBeTrue)
		})

		Convey("SetSliceValue with pointer to slice", func() {
			var is *[]int
			err := SetSliceValue(reflect.ValueOf(&is).Elem(), []string{"1", "2"})
			So(err, ShouldBeNil)
			So(*is, ShouldResemble, []int{1, 2})
		})

		Convey("SetPrimitiveValue with unsigned integer errors", func() {
			var u uint
			err := SetPrimitiveValue(reflect.ValueOf(&u).Elem(), "abc")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTypeMismatch), ShouldBeTrue)
		})

		Convey("SetPrimitiveValue with float errors", func() {
			var f float64
			err := SetPrimitiveValue(reflect.ValueOf(&f).Elem(), "abc")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTypeMismatch), ShouldBeTrue)
		})

		Convey("SetPrimitiveValue with bool errors", func() {
			var b bool
			err := SetPrimitiveValue(reflect.ValueOf(&b).Elem(), "abc")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTypeMismatch), ShouldBeTrue)
		})

		Convey("ExtractStructTags with pointer type", func() {
			type Simple struct {
				Foo string `query:"f"`
			}
			p, _, _, _ := ExtractStructTags(reflect.TypeOf(&Simple{}))
			So(p, ShouldBeEmpty) // Simple doesn't have path tag
		})

		Convey("ExtractStructTags with unexported fields", func() {
			type Unexported struct {
				foo string `query:"foo"`
			}
			p, q, h, b := ExtractStructTags(reflect.TypeOf(Unexported{}))
			So(p, ShouldBeEmpty)
			So(q, ShouldBeEmpty)
			So(h, ShouldBeEmpty)
			So(b, ShouldBeEmpty)
		})
	})
}
