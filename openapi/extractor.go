package openapi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ExtractorContext tracks the extraction state to prevent infinite recursion
// and manage shared $ref schemas.
type ExtractorContext struct {
	Document *openapi3.T
	Depth    int
}

// ErrorResponse represents the globally defined structure for non-2xx status codes.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// RouteMetadata holds human-readable information to annotate the OpenAPI operation.
type RouteMetadata struct {
	Tags        []string
	Summary     string
	Description string
}

func extractRequest(t reflect.Type, op *openapi3.Operation, ctx *ExtractorContext) {
	var bodyFields []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		isParam := false
		paramIn := ""
		paramName := ""

		if tag := field.Tag.Get("path"); tag != "" {
			isParam = true
			paramIn = "path"
			paramName = tag
		} else if tag := field.Tag.Get("query"); tag != "" {
			isParam = true
			paramIn = "query"
			paramName = tag
		} else if tag := field.Tag.Get("header"); tag != "" {
			isParam = true
			paramIn = "header"
			paramName = tag
		}

		if isParam {
			param := &openapi3.Parameter{}
			param.Name = paramName
			param.In = paramIn
			param.Schema = getOrAddSchema(field.Type, ctx)
			if paramIn == "path" {
				param.Required = true
			}
			op.AddParameter(param)
		} else {
			// Specific body field or any other field goes to body
			if strings.ToLower(field.Name) == "body" {
				schemaRef := getOrAddSchema(field.Type, ctx)
				op.RequestBody = &openapi3.RequestBodyRef{
					Value: openapi3.NewRequestBody().WithJSONSchemaRef(schemaRef),
				}
			} else {
				bodyFields = append(bodyFields, field)
			}
		}
	}

	// If no explicit 'Body' field was found, map other fields into a JSON schema (inline)
	if op.RequestBody == nil && len(bodyFields) > 0 {
		schema := openapi3.NewObjectSchema()
		for _, f := range bodyFields {
			name := f.Tag.Get("json")
			if name == "" {
				name = f.Name
			}
			schema.Properties[name] = getOrAddSchema(f.Type, ctx)
		}
		op.RequestBody = &openapi3.RequestBodyRef{
			Value: openapi3.NewRequestBody().WithJSONSchemaRef(&openapi3.SchemaRef{Value: schema}),
		}
	}
}

func extractResponse(t reflect.Type, successStatus int, op *openapi3.Operation, ctx *ExtractorContext) {
	if successStatus <= 0 {
		successStatus = 200
	}
	status := "200" // default success status

	resp := openapi3.NewResponse().WithDescription("Success response")
	resp.Headers = make(openapi3.Headers)

	var bodyField *reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.ToLower(field.Name) == "status" {
			// Status field is just metadata for the library, we can't extract the exact int value statically easily here.
			// But we could use it if we needed to. For now, assuming default 200.
			continue
		} else if strings.ToLower(field.Name) == "body" {
			bodyField = &field
		} else if tag := field.Tag.Get("header"); tag != "" {
			header := &openapi3.Header{
				Parameter: openapi3.Parameter{
					Schema: getOrAddSchema(field.Type, ctx),
				},
			}
			resp.Headers[tag] = &openapi3.HeaderRef{Value: header}
		}
	}

	if bodyField != nil {
		resp.Content = openapi3.NewContentWithJSONSchemaRef(getOrAddSchema(bodyField.Type, ctx))
	}

	op.Responses.Set(status, &openapi3.ResponseRef{Value: resp})
}

func getOrAddSchema(t reflect.Type, ctx *ExtractorContext) *openapi3.SchemaRef {
	if ctx.Depth > 10 {
		// Stop recursion
		return &openapi3.SchemaRef{Value: openapi3.NewObjectSchema()}
	}

	ctx.Depth++
	defer func() { ctx.Depth-- }()

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &openapi3.SchemaRef{Value: openapi3.NewIntegerSchema()}
	case reflect.Float32, reflect.Float64:
		return &openapi3.SchemaRef{Value: openapi3.NewFloat64Schema()}
	case reflect.Bool:
		return &openapi3.SchemaRef{Value: openapi3.NewBoolSchema()}
	case reflect.Slice, reflect.Array:
		items := getOrAddSchema(t.Elem(), ctx)
		schema := openapi3.NewArraySchema()
		schema.Items = items
		return &openapi3.SchemaRef{Value: schema}
	case reflect.Map:
		schema := openapi3.NewObjectSchema()
		schema.AdditionalProperties.Schema = getOrAddSchema(t.Elem(), ctx)
		return &openapi3.SchemaRef{Value: schema}
	case reflect.Struct:
		name := t.Name()
		if name == "" {
			// Anonymous struct
			schema := openapi3.NewObjectSchema()
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				jsonTag := f.Tag.Get("json")
				if jsonTag == "-" {
					continue
				}
				if jsonTag == "" {
					jsonTag = f.Name
				} else {
					jsonTag = strings.Split(jsonTag, ",")[0]
				}
				schema.Properties[jsonTag] = getOrAddSchema(f.Type, ctx)
			}
			return &openapi3.SchemaRef{Value: schema}
		}

		// Named struct - use components
		ref := fmt.Sprintf("#/components/schemas/%s", name)
		if ctx.Document.Components.Schemas == nil {
			ctx.Document.Components.Schemas = make(openapi3.Schemas)
		}

		if _, exists := ctx.Document.Components.Schemas[name]; !exists {
			// Pre-declare to prevent infinite recursion on self-referential types
			schema := openapi3.NewObjectSchema()
			ctx.Document.Components.Schemas[name] = &openapi3.SchemaRef{Value: schema}

			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				jsonTag := f.Tag.Get("json")
				if jsonTag == "-" {
					continue
				}
				if jsonTag == "" {
					jsonTag = f.Name
				} else {
					jsonTag = strings.Split(jsonTag, ",")[0]
				}
				schema.Properties[jsonTag] = getOrAddSchema(f.Type, ctx)
			}
		}
		return &openapi3.SchemaRef{Ref: ref}
	case reflect.Interface:
		return &openapi3.SchemaRef{Value: openapi3.NewSchema()} // any type
	default:
		return &openapi3.SchemaRef{Value: openapi3.NewSchema()}
	}
}
