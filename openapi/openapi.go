package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/route"
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPI represents the OpenAPI specification for a mini API.
type OpenAPI struct {
	spec    *openapi3.T
	adapter miniapi.Adapter
	docPath string
	uiPath  string
	uiLib   UILib
}

// SetSpec implements miniapi.OpenAPI.
func (o *OpenAPI) SetSpec(spec *openapi3.T) {
	o.spec = spec
}

// Spec implements miniapi.OpenAPI.
func (o *OpenAPI) Spec() *openapi3.T {
	return o.spec
}

// Build implements miniapi.OpenAPI.
func (o *OpenAPI) Build(group miniapi.Group) error {
	info := o.spec.Info
	if info == nil {
		info = &openapi3.Info{
			Title:   "Mini API",
			Version: "1.0.0",
		}
		o.spec.Info = info
	}

	if err := o.Attach("", group, o.adapter, nil, nil); err != nil {
		return err
	}

	docData, err := json.Marshal(o.spec)
	if err != nil {
		return err
	}

	docPath := o.docPath
	if docPath == "" {
		docPath = "/docs/openapi.json"
	}

	dRoute := route.NewWithHandlerFunc(route.Spec{
		Path:    docPath,
		Method:  http.MethodGet,
		Tags:    []string{"Documentation"},
		Summary: "Get OpenAPI Specification",
	}, BuildDocsHandler(docData))
	o.adapter.AddRoute("", dRoute)

	uiPath := o.uiPath
	if uiPath == "" {
		uiPath = "/docs/ui"
	}

	uRoute := route.NewWithHandlerFunc(route.Spec{
		Path:        uiPath,
		Method:      http.MethodGet,
		Tags:        []string{"Documentation"},
		Summary:     "Get API Documentation UI",
		Description: "Returns the API documentation UI for the API.",
	}, BuildUIHandler(docPath, o.uiLib))
	o.adapter.AddRoute("", uRoute)
	return nil
}

// Attach implements miniapi.OpenAPI.
func (o *OpenAPI) Attach(prefix string, group miniapi.Group, adapter miniapi.Adapter, parentModifiers []miniapi.RouteModifier, parentMiddlewares []miniapi.Middleware) error {
	var (
		extendedModifiers   = append(parentModifiers, group.Modifiers()...)
		extendedMiddlewares = append(parentMiddlewares, group.Middlewares()...)
	)
	for route := range group.Routes() {
		// modify route with modifiers.
		var modifiers []miniapi.RouteModifier
		modifiers = append(modifiers, extendedModifiers...)
		modifiers = append(modifiers, route.Modifiers()...)
		for _, modifier := range modifiers {
			route = modifier(route)
		}

		// Apply middlewares.
		var middlewares []miniapi.Middleware
		middlewares = append(middlewares, extendedMiddlewares...)
		middlewares = append(middlewares, route.Middlewares()...)
		handlerFunc := route.HandlerFunc()
		for i := len(middlewares) - 1; i >= 0; i-- {
			handlerFunc = middlewares[i](handlerFunc)
		}
		o.adapter.AddRoute(prefix+group.Path(), route)

		// build OpenAPI spec for the route.
		secReqs := o.spec.Security
		if len(group.Security()) > 0 {
			secReqs = ParseSecurityRequirements(group.Security())
		}
		if err := o.setOperationToSpec(prefix+group.Path(), route, secReqs); err != nil {
			return err
		}
	}

	for subgroup := range group.Groups() {
		if err := o.Attach(prefix+group.Path(), subgroup, adapter, extendedModifiers, extendedMiddlewares); err != nil {
			return err
		}
	}

	return nil
}

// setOperationToSpec sets the OpenAPI operation to the spec for the given route.
func (o *OpenAPI) setOperationToSpec(prefix string, route miniapi.Route, secReqs openapi3.SecurityRequirements) error {
	ctx := &ExtractorContext{
		Document: o.spec,
		Depth:    0,
	}

	meta := route.Operation()
	if len(meta.Security) > 0 {
		secReqs = ParseSecurityRequirements(meta.Security)
	}

	op := openapi3.NewOperation()
	op.Tags = meta.Tags
	op.Summary = meta.Summary
	op.Description = meta.Description
	op.Deprecated = meta.Deprecated
	op.Responses = openapi3.NewResponses()
	if len(secReqs) > 0 {
		op.Security = &secReqs
		for _, secReq := range secReqs {
			for scheme := range secReq {
				if o.spec.Components.SecuritySchemes == nil {
					o.spec.Components.SecuritySchemes = make(openapi3.SecuritySchemes)
				}
				if _, exists := o.spec.Components.SecuritySchemes[scheme]; !exists {
					o.spec.Components.SecuritySchemes[scheme] = &openapi3.SecuritySchemeRef{
						Value: openapi3.NewSecurityScheme().WithType("http").WithScheme("Bearer"),
					}
				}
			}
		}
	}

	var statuses []int
	for s := range meta.ResponseTypes {
		statuses = append(statuses, s)
	}
	sort.Ints(statuses)

	setDefault := false
	for _, status := range statuses {
		resp := meta.ResponseTypes[status]

		description := resp.Description
		if description == "" {
			description = http.StatusText(status)
		}

		response := openapi3.NewResponse().
			WithDescription(description)

		if resp.Schema != nil {
			schemaRef := getOrAddSchema(resp.Schema, ctx)

			contentType := resp.ContentType
			if contentType == "" {
				contentType = "application/json"
			}

			response.Content = openapi3.Content{
				resp.ContentType: &openapi3.MediaType{
					Schema: schemaRef,
				},
			}
		}

		op.Responses.Set(
			strconv.Itoa(status),
			&openapi3.ResponseRef{Value: response},
		)

		if !setDefault && resp.Default {
			op.Responses.Set("default", &openapi3.ResponseRef{Value: response})
			setDefault = true
		}
	}

	reqType := route.ReqType()
	if reqType != nil {
		if reqType.Kind() == reflect.Pointer {
			reqType = reqType.Elem()
		}
		if reqType.Kind() == reflect.Struct {
			extractRequest(reqType, op, ctx)
		}
	}
	resType := route.ResType()
	if resType != nil {
		if resType.Kind() == reflect.Pointer {
			resType = resType.Elem()
		}
		if resType.Kind() == reflect.Struct {
			extractResponse(resType, meta.DefaultStatus, op, ctx)
		}
	}

	var (
		fullPath = prefix + route.Path()
		method   = strings.ToUpper(route.Method())
	)
	pathItem := o.spec.Paths.Value(fullPath)
	if pathItem == nil {
		pathItem = &openapi3.PathItem{}
		o.spec.Paths.Set(fullPath, pathItem)
	}
	pathItem.SetOperation(method, op)
	return nil
}

var _ miniapi.OpenAPI = (*OpenAPI)(nil)

// UILib represents the UI library used for rendering the API documentation UI.
// ENUM(redoc,scalar)
//
//go:generate go-enum
type UILib string

// Config represents the configuration for the OpenAPI specification.
type Config struct {
	UILib   UILib
	Spec    *openapi3.T
	Adapter miniapi.Adapter
	DocPath string
	UIPath  string
}

// New creates a new OpenAPI specification for a mini API.
func New(cfg Config) *OpenAPI {
	spec := cfg.Spec
	if spec.OpenAPI == "" {
		spec.OpenAPI = "3.0.5"
	}
	if spec.Paths == nil {
		spec.Paths = openapi3.NewPaths()
	}
	if spec.Info == nil {
		spec.Info = &openapi3.Info{
			Title:   "Mini API",
			Version: "1.0.0",
		}
	}
	if spec.Components == nil {
		spec.Components = &openapi3.Components{}
	}
	var server *openapi3.Server
	for _, srv := range spec.Servers {
		if srv.URL == cfg.Adapter.Address() {
			server = srv
		}
	}
	if server == nil {
		spec.Servers = append(spec.Servers, &openapi3.Server{
			URL:         cfg.Adapter.Address(),
			Description: fmt.Sprintf("Server for %s", spec.Info.Title),
			Variables:   map[string]*openapi3.ServerVariable{},
		})
	}
	return &OpenAPI{
		spec:    spec,
		adapter: cfg.Adapter,
		docPath: cfg.DocPath,
		uiPath:  cfg.UIPath,
		uiLib:   cfg.UILib,
	}
}

// ParseSecurityRequirements converts a slice of miniapi.SecurityRequirement to
// an openapi3.SecurityRequirements.
func ParseSecurityRequirements(secs []miniapi.SecurityRequirement) openapi3.SecurityRequirements {
	if len(secs) == 0 {
		return make(openapi3.SecurityRequirements, 0)
	}
	res := openapi3.NewSecurityRequirements()
	for _, sec := range secs {
		res.With(openapi3.SecurityRequirement(sec))
	}
	return *res
}
