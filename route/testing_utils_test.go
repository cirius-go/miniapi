package route

import (
	"errors"
)

type MockRequest struct {
	ID string `json:"id" path:"id"`
}

type MockResponse struct {
	Result string `json:"result"`
}

type MockError struct {
	Message string `json:"message"`
}

var ErrMockBusinessLogic = errors.New("mock business logic error")
