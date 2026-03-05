package group

import (
	"github.com/cirius-go/miniapi"
)

// Group represents the group of routes in miniapi.
type Group struct {
	prefix       string
	binder       miniapi.Binder
	errorEncoder miniapi.ErrorEncoder
	routes       []miniapi.RouteBuilder
	groups       []miniapi.Group
	modifiers    []miniapi.Modifier
	middlewares  []miniapi.Middleware
}

// Build implements miniapi.Group.
func (g *Group) Build(ctx miniapi.BuildContext) []miniapi.Route {
	ctx = ctx.Clone()

	ctx.Modifiers = append(ctx.Modifiers, g.modifiers...)
	ctx.Middlewares = append(ctx.Middlewares, g.middlewares...)
	ctx.Prefix += g.prefix

	if g.binder != nil {
		ctx.Binder = g.binder
	}
	if g.errorEncoder != nil {
		ctx.ErrorEncoder = g.errorEncoder
	}

	var compiledRoutes []miniapi.Route

	for _, route := range g.routes {
		compiledRoutes = append(compiledRoutes, route.Build(ctx))
	}

	for _, group := range g.groups {
		compiledRoutes = append(compiledRoutes, group.Build(ctx)...)
	}

	return compiledRoutes
}

// Binder implements miniapi.Group.
func (g *Group) Binder() miniapi.Binder {
	return g.binder
}

// ErrorEncoder implements miniapi.Group.
func (g *Group) ErrorEncoder() miniapi.ErrorEncoder {
	return g.errorEncoder
}

// SetBinder implements miniapi.Group.
func (g *Group) SetBinder(b miniapi.Binder) miniapi.Group {
	g.binder = b
	return g
}

// SetErrorEncoder implements miniapi.Group.
func (g *Group) SetErrorEncoder(e miniapi.ErrorEncoder) miniapi.Group {
	g.errorEncoder = e
	return g
}

// SetPrefix implements miniapi.Group.
func (g *Group) SetPrefix(prefix string) miniapi.Group {
	g.prefix = prefix
	return g
}

// AddMiddlewares implements miniapi.Group.
func (g *Group) AddMiddlewares(middlewares ...miniapi.Middleware) miniapi.Group {
	g.middlewares = append(g.middlewares, middlewares...)
	return g
}

// AddModifiers implements miniapi.Group.
func (g *Group) AddModifiers(modifiers ...miniapi.Modifier) miniapi.Group {
	g.modifiers = append(g.modifiers, modifiers...)
	return g
}

// AddRoutes implements miniapi.Group.
func (g *Group) AddRoutes(routes ...miniapi.RouteBuilder) miniapi.Group {
	g.routes = append(g.routes, routes...)
	return g
}

// AddGroups implements miniapi.Group.
func (g *Group) AddGroups(groups ...miniapi.Group) miniapi.Group {
	g.groups = append(g.groups, groups...)
	return g
}

// Middlewares implements miniapi.Group.
func (g *Group) Middlewares() []miniapi.Middleware {
	return g.middlewares
}

// Modifiers implements miniapi.Group.
func (g *Group) Modifiers() []miniapi.Modifier {
	return g.modifiers
}

// NewGroup implements miniapi.Group.
func (g *Group) NewGroup(prefix string) miniapi.Group {
	child := New(prefix)

	g.groups = append(g.groups, child)
	return child
}

// Prefix implements miniapi.Group.
func (g *Group) Prefix() string {
	return g.prefix
}

var _ miniapi.Group = (*Group)(nil)

// Spec represents the specification of the group.
type Spec struct {
	Binder       miniapi.Binder
	ErrorEncoder miniapi.ErrorEncoder
	Routes       []miniapi.RouteBuilder
	Groups       []miniapi.Group
	Modifiers    []miniapi.Modifier
	Middlewares  []miniapi.Middleware
}

// New creates a new Group.
func New(prefix string, specs ...Spec) *Group {
	var spec Spec
	if len(specs) > 0 {
		spec = specs[0]
	}

	return &Group{
		binder:       spec.Binder,
		errorEncoder: spec.ErrorEncoder,
		prefix:       prefix,
		modifiers:    append([]miniapi.Modifier{}, spec.Modifiers...),
		middlewares:  append([]miniapi.Middleware{}, spec.Middlewares...),
		routes:       append([]miniapi.RouteBuilder{}, spec.Routes...),
		groups:       append([]miniapi.Group{}, spec.Groups...),
	}
}
