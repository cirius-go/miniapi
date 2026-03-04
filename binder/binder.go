package binder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/cirius-go/miniapi"
)

// BindingOptions configures the behavior of the binding engine.
type BindingOptions struct {
	MaxBodySize           int64
	DisallowUnknownFields bool
}

// DefaultBindingOptions provides sensible defaults for the binding engine.
var DefaultBindingOptions = BindingOptions{
	MaxBodySize:           1024 * 1024, // 1MB default
	DisallowUnknownFields: false,
}

// DefaultBinder is the standard implementation of the miniapi.Binder interface.
type DefaultBinder struct {
	Options BindingOptions
}

// New returns a new DefaultBinder with the given options.
func New(opts BindingOptions) *DefaultBinder {
	return &DefaultBinder{Options: opts}
}

// BindRequest parses path, query, header, and body parameters into the given struct req.
func (b *DefaultBinder) BindRequest(ctx miniapi.Context, req any) error {
	rv := reflect.ValueOf(req)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("%w: req must be a pointer to a struct", ErrBindingFailed)
	}

	rv = rv.Elem()
	pathMeta, queryMeta, headerMeta, bodyIndex := extractStructTags(rv.Type())

	if err := bindPathParams(ctx, rv, pathMeta); err != nil {
		return err
	}
	if err := bindQueryParams(ctx, rv, queryMeta); err != nil {
		return err
	}
	if err := bindHeaderParams(ctx, rv, headerMeta); err != nil {
		return err
	}
	if len(bodyIndex) > 0 {
		if err := bindBody(ctx, rv, bodyIndex, b.Options); err != nil {
			return err
		}
	}

	return nil
}

// bindPathParams extracts path parameters and sets them in the struct.
func bindPathParams(ctx miniapi.Context, rv reflect.Value, meta []StructFieldMeta) error {
	for _, m := range meta {
		val := ctx.RequestParam(m.Tag)
		if val == "" {
			continue // Mandatory validation is deferred to external packages
		}

		field := rv.FieldByIndex(m.FieldIndex)
		if err := setPrimitiveValue(field, val); err != nil {
			return err
		}
	}
	return nil
}

// bindQueryParams extracts query parameters and sets them in the struct.
func bindQueryParams(ctx miniapi.Context, rv reflect.Value, meta []StructFieldMeta) error {
	u := ctx.RequestURL()
	queryValues := u.Query()

	for _, m := range meta {
		vals, ok := queryValues[m.Tag]
		if !ok || len(vals) == 0 {
			continue
		}

		field := rv.FieldByIndex(m.FieldIndex)
		if m.IsSlice {
			if err := setSliceValue(field, vals); err != nil {
				return err
			}
		} else {
			if err := setPrimitiveValue(field, vals[0]); err != nil {
				return err
			}
		}
	}
	return nil
}

// bindHeaderParams extracts header parameters and sets them in the struct.
func bindHeaderParams(ctx miniapi.Context, rv reflect.Value, meta []StructFieldMeta) error {
	for _, m := range meta {
		var vals []string

		// Collect all header values for the tag
		for k, v := range ctx.IterRequestHeader() {
			if strings.EqualFold(k, m.Tag) {
				vals = append(vals, v)
			}
		}

		if len(vals) == 0 {
			continue
		}

		field := rv.FieldByIndex(m.FieldIndex)
		if m.IsSlice {
			if err := setSliceValue(field, vals); err != nil {
				return err
			}
		} else {
			if err := setPrimitiveValue(field, vals[0]); err != nil {
				return err
			}
		}
	}
	return nil
}

// bindBody reads the request body and sets it in the body field.
func bindBody(ctx miniapi.Context, rv reflect.Value, bodyIndex []int, opts BindingOptions) error {
	field := rv.FieldByIndex(bodyIndex)

	// If it's a pointer, allocate it if nil
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
	}

	reader := ctx.RequestBodyReader()
	if reader == nil {
		return nil
	}
	defer reader.Close()

	if opts.MaxBodySize > 0 {
		reader = http.MaxBytesReader(nil, reader, opts.MaxBodySize)
	}

	// Handle io.Reader
	if field.Type() == reflect.TypeOf((*io.Reader)(nil)).Elem() {
		// Read all bytes to memory so that the field holds the data,
		// because the request reader might be closed after this function returns.
		// Alternatively, we could just assign a buffer.
		// Wait, if it's an io.Reader, giving it a bytes.Buffer or strings.Reader is safer.
		b, err := io.ReadAll(reader)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				return fmt.Errorf("%w: %s", ErrPayloadTooLarge, err.Error())
			}
			return fmt.Errorf("%w: %s", ErrBindingFailed, err.Error())
		}
		field.Set(reflect.ValueOf(strings.NewReader(string(b))))
		return nil
	}

	// Handle []byte
	if field.Type() == reflect.TypeOf([]byte{}) {
		b, err := io.ReadAll(reader)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				return fmt.Errorf("%w: %s", ErrPayloadTooLarge, err.Error())
			}
			return fmt.Errorf("%w: %s", ErrBindingFailed, err.Error())
		}
		field.Set(reflect.ValueOf(b))
		return nil
	}

	// For structs, maps, etc., check Content-Type
	contentType := ctx.RequestHeader("Content-Type")
	// Clean up content type (e.g. 'application/json; charset=utf-8' -> 'application/json')
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	contentType = strings.ToLower(contentType)

	if contentType != "" && contentType != "application/json" {
		return fmt.Errorf("%w: %s", ErrUnsupportedMediaType, contentType)
	}

	// Use JSON unmarshaling (default and currently only supported format for structs)
	decoder := json.NewDecoder(reader)
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(field.Addr().Interface()); err != nil && err != io.EOF {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return fmt.Errorf("%w: %s", ErrPayloadTooLarge, err.Error())
		}
		return fmt.Errorf("%w: failed to decode request body: %s", ErrBindingFailed, err.Error())
	}

	return nil
}

// MarshalResponse marshals the given response struct into the HTTP response writer
// based on the Accept header.
func (b *DefaultBinder) MarshalResponse(ctx miniapi.Context, res any) error {
	if res == nil {
		return nil
	}

	accept := ctx.RequestHeader("Accept")
	// Simple content negotiation
	isPlainText := false
	if strings.Contains(strings.ToLower(accept), "text/plain") {
		// Verify text/plain is preferred over application/json
		// This is a naive implementation; a full q-value parser is more complex,
		// but this serves MVP needs.
		jsonIdx := strings.Index(strings.ToLower(accept), "application/json")
		textIdx := strings.Index(strings.ToLower(accept), "text/plain")
		if jsonIdx == -1 || textIdx < jsonIdx {
			isPlainText = true
		}
	}

	if isPlainText {
		if ctx.ResponseHeader("Content-Type") == "" {
			ctx.SetResponseHeader("Content-Type", "text/plain")
		}

		// For text/plain, we write the string representation or just use fmt.Sprintf
		var b []byte
		if s, ok := res.(string); ok {
			b = []byte(s)
		} else if s, ok := res.(*string); ok && s != nil {
			b = []byte(*s)
		} else {
			b = fmt.Appendf(nil, "%v", res)
		}

		_, err := ctx.ResponseBodyWriter().Write(b)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrBindingFailed, err.Error())
		}
		return nil
	}

	// Default to JSON
	if ctx.ResponseHeader("Content-Type") == "" {
		ctx.SetResponseHeader("Content-Type", "application/json")
	}

	if err := json.NewEncoder(ctx.ResponseBodyWriter()).Encode(res); err != nil {
		return fmt.Errorf("%w: %s", ErrBindingFailed, err.Error())
	}
	return nil
}

// setPrimitiveValue parses the string value into the target primitive reflect.Value.
func setPrimitiveValue(v reflect.Value, value string) error {
	// If it's a pointer, allocate and dereference it
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("%w: cannot parse %q as %s", ErrTypeMismatch, value, v.Kind().String())
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("%w: cannot parse %q as %s", ErrTypeMismatch, value, v.Kind().String())
		}
		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("%w: cannot parse %q as %s", ErrTypeMismatch, value, v.Kind().String())
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%w: cannot parse %q as %s", ErrTypeMismatch, value, v.Kind().String())
		}
		v.SetBool(b)
	default:
		return fmt.Errorf("%w: unsupported primitive kind %s", ErrTypeMismatch, v.Kind().String())
	}
	return nil
}

// setSliceValue parses a slice of strings into a target slice reflect.Value.
func setSliceValue(v reflect.Value, values []string) error {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		return fmt.Errorf("%w: target is not a slice", ErrTypeMismatch)
	}

	slice := reflect.MakeSlice(v.Type(), len(values), len(values))
	for i, val := range values {
		elem := slice.Index(i)
		if err := setPrimitiveValue(elem, val); err != nil {
			return err
		}
	}
	v.Set(slice)
	return nil
}

// StructFieldMeta holds metadata about a field for binding.
type StructFieldMeta struct {
	Name       string
	Tag        string
	Value      string
	IsSlice    bool
	FieldIndex []int
}

// extractStructTags inspects the struct type and returns metadata for binding path, query, and header.
// It also identifies the field index of the "Body" field if it exists.
func extractStructTags(t reflect.Type) (pathMeta, queryMeta, headerMeta []StructFieldMeta, bodyIndex []int) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Unexported fields
		if field.PkgPath != "" {
			continue
		}

		if field.Name == "Body" {
			bodyIndex = []int{i}
		}

		if pathTag := field.Tag.Get("path"); pathTag != "" {
			pathMeta = append(pathMeta, StructFieldMeta{Name: field.Name, Tag: pathTag, IsSlice: field.Type.Kind() == reflect.Slice, FieldIndex: []int{i}})
		}
		if queryTag := field.Tag.Get("query"); queryTag != "" {
			queryMeta = append(queryMeta, StructFieldMeta{Name: field.Name, Tag: queryTag, IsSlice: field.Type.Kind() == reflect.Slice, FieldIndex: []int{i}})
		}
		if headerTag := field.Tag.Get("header"); headerTag != "" {
			headerMeta = append(headerMeta, StructFieldMeta{Name: field.Name, Tag: headerTag, IsSlice: field.Type.Kind() == reflect.Slice, FieldIndex: []int{i}})
		}
	}
	return
}
