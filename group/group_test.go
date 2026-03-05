package group

import (
	"testing"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestGroup(t *testing.T) {
	Convey("Given a Group instance", t, func() {
		g := New("/api")

		Convey("When checking the prefix", func() {
			So(g.Prefix(), ShouldEqual, "/api")
		})

		Convey("When setting a new prefix", func() {
			g.SetPrefix("/v1")
			So(g.Prefix(), ShouldEqual, "/v1")
		})

		Convey("When creating a child group using NewGroup", func() {
			child := g.NewGroup("/users")
			So(child.Prefix(), ShouldEqual, "/users")
			
			Convey("It should be added to the parent's groups list", func() {
				So(len(g.groups), ShouldEqual, 1)
				So(g.groups[0], ShouldEqual, child)
			})
		})

		Convey("When adding routes to the group", func() {
			mockRoute := new(mocks.RouteBuilder)
			g.AddRoutes(mockRoute)
			So(len(g.routes), ShouldEqual, 1)
			So(g.routes[0], ShouldEqual, mockRoute)
		})

		Convey("When adding multiple groups at once", func() {
			g1 := New("/1")
			g2 := New("/2")
			g.AddGroups(g1, g2)
			So(len(g.groups), ShouldEqual, 2)
		})

		Convey("When setting and adding configurations", func() {
			mockBinder := new(mocks.Binder)
			mockEncoder := func(miniapi.Context, error) {}
			mockMiddleware := func(miniapi.HandlerFunc) miniapi.HandlerFunc { return nil }
			mockModifier := func(miniapi.RouteBuilder) miniapi.RouteBuilder { return nil }

			g.SetBinder(mockBinder)
			g.SetErrorEncoder(mockEncoder)
			g.AddMiddlewares(mockMiddleware)
			g.AddModifiers(mockModifier)

			So(g.Binder(), ShouldEqual, mockBinder)
			So(g.ErrorEncoder(), ShouldNotBeNil) // Functions can't be compared directly easily
			So(len(g.Middlewares()), ShouldEqual, 1)
			So(len(g.Modifiers()), ShouldEqual, 1)
		})

		Convey("When building the group", func() {
			child := g.NewGroup("/v1")
			
			mockRoute := new(mocks.RouteBuilder)
			child.AddRoutes(mockRoute)

			Convey("It should correctly concatenate prefixes and propagate configurations", func() {
				// We need a dummy route to return from Build
				dummyRoute := new(mocks.Route)
				
				// Mock expectations for the route builder
				mockRoute.On("Build", mock.MatchedBy(func(ctx miniapi.BuildContext) bool {
					return ctx.Prefix == "/api/v1"
				})).Return(dummyRoute)

				routes := g.Build(miniapi.BuildContext{})
				
				So(len(routes), ShouldEqual, 1)
				So(routes[0], ShouldEqual, dummyRoute)
				mockRoute.AssertExpectations(t)
			})

			Convey("It should respect overrides in the BuildContext", func() {
				mockBinder := new(mocks.Binder)
				mockEncoder := func(miniapi.Context, error) {}
				g.SetBinder(mockBinder)
				g.SetErrorEncoder(mockEncoder)

				dummyRoute := new(mocks.Route)
				mockRoute.On("Build", mock.MatchedBy(func(ctx miniapi.BuildContext) bool {
					return ctx.Binder == mockBinder && ctx.ErrorEncoder != nil
				})).Return(dummyRoute)

				routes := g.Build(miniapi.BuildContext{})
				So(len(routes), ShouldEqual, 1)
				So(routes[0], ShouldEqual, dummyRoute)
				mockRoute.AssertExpectations(t)
			})
		})

		Convey("When creating a group with Spec", func() {
			mockBinder := new(mocks.Binder)
			spec := Spec{
				Binder: mockBinder,
			}
			g2 := New("/spec", spec)
			So(g2.Binder(), ShouldEqual, mockBinder)
			So(g2.Prefix(), ShouldEqual, "/spec")
		})

		Convey("When handling edge cases", func() {
			Convey("Empty path segments", func() {
				g.SetPrefix("")
				child := g.NewGroup("")
				So(g.Prefix(), ShouldEqual, "")
				So(child.Prefix(), ShouldEqual, "")
				
				dummyRoute := new(mocks.Route)
				mockRoute := new(mocks.RouteBuilder)
				mockRoute.On("Build", mock.MatchedBy(func(ctx miniapi.BuildContext) bool {
					return ctx.Prefix == ""
				})).Return(dummyRoute)
				
				g.AddRoutes(mockRoute)
				routes := g.Build(miniapi.BuildContext{})
				So(len(routes), ShouldEqual, 1)
			})

			Convey("Nil routes, modifiers, or middlewares", func() {
				// The implementation doesn't check for nil in AddRoutes/AddMiddlewares/AddModifiers
				// but Build loops over them. If route is nil, it will panic.
				// However, middlewares and modifiers are functions, so we should check if they are nil before calling them.
				// Currently, Group just appends them.
				
				So(func() { g.AddRoutes(nil) }, ShouldNotPanic)
				So(func() { g.AddMiddlewares(nil) }, ShouldNotPanic)
				So(func() { g.AddModifiers(nil) }, ShouldNotPanic)
			})
		})
	})
}
