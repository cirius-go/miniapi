package route

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResponseSchema(t *testing.T) {
	Convey("Given Response Schema Builder Functions", t, func() {
		Convey("When calling JSONSchema", func() {
			res := JSONSchema[MockResponse]("A successful response")

			Convey("It should set the content type to application/json", func() {
				So(res.ContentType, ShouldEqual, "application/json")
			})

			Convey("It should correctly assign the description", func() {
				So(res.Description, ShouldEqual, "A successful response")
			})

			Convey("It should correctly reflect the schema type", func() {
				So(res.Schema, ShouldEqual, reflect.TypeOf(MockResponse{}))
			})

			Convey("It should not be marked as NoContent", func() {
				So(res.NoContent, ShouldBeFalse)
			})
		})

		Convey("When calling ProblemSchema", func() {
			res := ProblemSchema[MockError]("An error occurred")

			Convey("It should set the content type to application/problem+json", func() {
				So(res.ContentType, ShouldEqual, "application/problem+json")
			})

			Convey("It should correctly assign the description", func() {
				So(res.Description, ShouldEqual, "An error occurred")
			})

			Convey("It should correctly reflect the schema type", func() {
				So(res.Schema, ShouldEqual, reflect.TypeOf(MockError{}))
			})

			Convey("It should not be marked as NoContent", func() {
				So(res.NoContent, ShouldBeFalse)
			})
		})

		Convey("When calling NoContentSchema", func() {
			res := NoContentSchema("No content here")

			Convey("It should correctly assign the description", func() {
				So(res.Description, ShouldEqual, "No content here")
			})

			Convey("It should have a nil schema", func() {
				So(res.Schema, ShouldBeNil)
			})

			// NOTE: Currently NoContentSchema does not explicitly set NoContent=true in the struct.
			// Depending on the expected behavior, this might need an update in the codebase.
			// We will test its current behavior based on the implementation in response_schema.go
			// which just returns Schema: nil.
		})
	})
}
