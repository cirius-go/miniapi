package group

import (
	"iter"

	"github.com/cirius-go/miniapi"
)

// Group represents the group of routes in miniapi.
type Group struct {
	path        string
	parent      *Group
	routes      map[string]miniapi.Route
	groups      map[string]miniapi.Group
	modifiers   []miniapi.RouteModifier
	security    []miniapi.SecurityRequirement
	middlewares []miniapi.Middleware
}

// AddMiddlewares implements miniapi.Group.
func (g *Group) AddMiddlewares(middlewares ...miniapi.Middleware) miniapi.Group {
	g.middlewares = append(g.middlewares, middlewares...)
	return g
}

// Middlewares implements miniapi.Group.
func (g *Group) Middlewares() []miniapi.Middleware {
	return g.middlewares
}

// SetMiddlewares implements miniapi.Group.
func (g *Group) SetMiddlewares(middlewares ...miniapi.Middleware) miniapi.Group {
	g.middlewares = middlewares
	return g
}

// WithSecurity overrides the global security for this group.
func (g *Group) WithSecurity(reqs ...miniapi.SecurityRequirement) miniapi.Group {
	g.security = reqs
	return g
}

// Security returns the explicitly set OpenAPI security requirements, or nil.
func (g *Group) Security() []miniapi.SecurityRequirement {
	return g.security
}

// AddModifiers implements miniapi.Group.
func (g *Group) AddModifiers(modifiers ...miniapi.RouteModifier) miniapi.Group {
	g.modifiers = append(g.modifiers, modifiers...)
	return g
}

// Modifiers implements miniapi.Group.
func (g *Group) Modifiers() []miniapi.RouteModifier {
	return g.modifiers
}

// Groups implements miniapi.Group.
func (g *Group) Groups() iter.Seq[miniapi.Group] {
	return func(yield func(miniapi.Group) bool) {
		for _, gr := range g.groups {
			if !yield(gr) {
				break
			}
		}
	}
}

// Routes implements miniapi.Group.
func (g *Group) Routes() iter.Seq[miniapi.Route] {
	return func(yield func(miniapi.Route) bool) {
		for _, r := range g.routes {
			if !yield(r) {
				break
			}
		}
	}
}

// FullPath returns the full path of the current group by concatenating the
// paths of all parent groups.
func (g *Group) FullPath() string {
	var (
		path    = ""
		current = g
	)
	for current != nil {
		path = current.Path() + path
		current = current.parent
	}
	return path
}

// NewGroup creates a new group with the given path and adds it to the current
// group.
// If a group with the same path already exists, it will return the existing
// group instead of creating a new one.
func (g *Group) NewGroup(path string) miniapi.Group {
	if gr, ok := g.groups[path]; ok {
		return gr
	}
	gr := New(path)
	gr.parent = g
	g.groups[path] = gr
	return gr
}

// AddRoutes adds the given routes to the current group.
// if a route with the same path and method already exists, it will skip adding
// new route.
func (g *Group) AddRoutes(routes ...miniapi.Route) miniapi.Group {
	for _, r := range routes {
		key := r.Path() + "__" + r.Method()
		if _, ok := g.routes[key]; ok {
			continue
		}
		g.routes[key] = r
	}
	return g
}

// Path implements *Group.
func (g *Group) Path() string {
	return g.path
}

// New creates a new Group.
func New(path string) *Group {
	return &Group{
		path:        path,
		parent:      nil,
		groups:      make(map[string]miniapi.Group),
		routes:      make(map[string]miniapi.Route),
		modifiers:   make([]miniapi.RouteModifier, 0),
		middlewares: make([]miniapi.Middleware, 0),
		security:    make([]miniapi.SecurityRequirement, 0),
	}
}
