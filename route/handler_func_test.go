package route

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/cirius-go/miniapi"
	"github.com/cirius-go/miniapi/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestMakeHandlerFuncBuilder(t *testing.T) {
	Convey("Given a TypedHandlerFunc and its Builder", t, func() {
		var userLogicCalled bool
		handler := func(ctx context.Context, req *MockRequest) (*MockResponse, error) {
			userLogicCalled = true
			if req.ID == "error" {
				return nil, ErrMockBusinessLogic
			}
			return &MockResponse{Result: "success"}, nil
		}

		builder := MakeHandlerFuncBuilder(handler)
		mockBinder := new(mocks.Binder)
		mockCtx := new(mocks.Context)
		mockCtx.On("RequestContext").Return(context.Background())

		Convey("When request binding and execution are successful", func() {
			mockBinder.On("BindRequest", mockCtx, mock.AnythingOfType("*route.MockRequest")).Return(nil)
			mockCtx.On("ResponseStatus").Return(0)
			mockCtx.On("SetResponseStatus", http.StatusOK).Return()
			mockBinder.On("MarshalResponse", mockCtx, mock.AnythingOfType("*route.MockResponse")).Return(nil)

			// Using a dummy error encoder that shouldn't be called
			dummyEncoder := func(c miniapi.Context, err error) {
				t.Error("Error encoder should not have been called")
			}

			fn := builder(mockBinder, dummyEncoder)
			fn(mockCtx)

			So(userLogicCalled, ShouldBeTrue)
			mockBinder.AssertExpectations(t)
			mockCtx.AssertExpectations(t)
		})

		Convey("When request binding fails", func() {
			expectedErr := errors.New("binding error")
			mockBinder.On("BindRequest", mockCtx, mock.AnythingOfType("*route.MockRequest")).Return(expectedErr)

			var encodedErr error
			encoder := func(c miniapi.Context, err error) {
				encodedErr = err
			}

			fn := builder(mockBinder, encoder)
			fn(mockCtx)

			So(userLogicCalled, ShouldBeFalse)
			So(encodedErr, ShouldEqual, expectedErr)
			mockBinder.AssertExpectations(t)
		})

		Convey("When user business logic returns an error", func() {
			// Trigger error in business logic
			mockBinder.On("BindRequest", mockCtx, mock.AnythingOfType("*route.MockRequest")).Run(func(args mock.Arguments) {
				req := args.Get(1).(*MockRequest)
				req.ID = "error"
			}).Return(nil)

			var encodedErr error
			encoder := func(c miniapi.Context, err error) {
				encodedErr = err
			}

			fn := builder(mockBinder, encoder)
			fn(mockCtx)

			So(userLogicCalled, ShouldBeTrue)
			So(encodedErr, ShouldEqual, ErrMockBusinessLogic)
			mockBinder.AssertExpectations(t)
		})

		Convey("When execution is successful but status is already set", func() {
			mockBinder.On("BindRequest", mockCtx, mock.AnythingOfType("*route.MockRequest")).Return(nil)
			mockCtx.On("ResponseStatus").Return(http.StatusCreated) // Status already set
			mockBinder.On("MarshalResponse", mockCtx, mock.AnythingOfType("*route.MockResponse")).Return(nil)

			dummyEncoder := func(c miniapi.Context, err error) {}

			fn := builder(mockBinder, dummyEncoder)
			fn(mockCtx)

			So(userLogicCalled, ShouldBeTrue)
			mockBinder.AssertExpectations(t)
			// Ensure SetResponseStatus was NOT called
			mockCtx.AssertNotCalled(t, "SetResponseStatus", mock.Anything)
		})
	})
}

type MockBuffer struct {
	buf []byte
}

func (m *MockBuffer) Write(p []byte) (n int, err error) {
	m.buf = append(m.buf, p...)
	return len(p), nil
}

func TestDefaultErrorEncoder(t *testing.T) {
	Convey("Given DefaultErrorEncoder", t, func() {
		mockCtx := new(mocks.Context)
		err := errors.New("test error")
		buffer := &MockBuffer{}

		Convey("It should set appropriate headers and status if not already set", func() {
			mockCtx.On("SetResponseHeader", "Content-Type", "application/problem+json").Return()
			mockCtx.On("ResponseStatus").Return(0)
			mockCtx.On("SetResponseStatus", http.StatusInternalServerError).Return()
			mockCtx.On("ResponseBodyWriter").Return(buffer)

			DefaultErrorEncoder(mockCtx, err)

			So(string(buffer.buf), ShouldContainSubstring, `"error":"test error"`)
			mockCtx.AssertExpectations(t)
		})

		Convey("It should not overwrite status if already set", func() {
			mockCtx.On("SetResponseHeader", "Content-Type", "application/problem+json").Return()
			mockCtx.On("ResponseStatus").Return(http.StatusBadRequest) // Already set
			// SetResponseStatus should NOT be called
			mockCtx.On("ResponseBodyWriter").Return(buffer)

			DefaultErrorEncoder(mockCtx, err)

			So(string(buffer.buf), ShouldContainSubstring, `"error":"test error"`)
			mockCtx.AssertExpectations(t)
			mockCtx.AssertNotCalled(t, "SetResponseStatus", mock.Anything)
		})
	})
}
