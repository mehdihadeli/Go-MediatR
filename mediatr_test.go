package mediatr

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_RegisterRequestHandler_Should_Return_Error_If_Handler_Already_Registered(t *testing.T) {
	expectedErr := fmt.Sprintf("registered handler already exists in the registry for message %s", "*mediatr.RequestTest")
	handler1 := &RequestTestHandler{}
	handler2 := &RequestTestHandler{}
	err1 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler1)
	err2 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler2)

	assert.Nil(t, err1)
	assert.Containsf(t, err2.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err2)
}

func Test_RegisterRequestHandler_Should_Register_All_Handlers(t *testing.T) {
	handler1 := &RequestTestHandler{}
	handler2 := &RequestTestHandler2{}
	err1 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler1)
	err2 := RegisterRequestHandler[*RequestTest2, *ResponseTest2](handler2)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func Test_Send_Should_Throw_Error_If_No_Handler_Registered(t *testing.T) {
	expectedErr := fmt.Sprintf("no handlers for command %T", &RequestTest{})
	_, err := Send[*RequestTest, *ResponseTest](context.Background(), &RequestTest{Data: "test"})
	assert.Containsf(t, err.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err)
}

func Test_Send_Should_Return_Error_If_Handler_Returns_Error(t *testing.T) {
	expectedErr := "error handling request"
	handler3 := &RequestTestHandler3{}
	_ = RegisterRequestHandler[*RequestTest2, *ResponseTest2](handler3)
	_, err := Send[*RequestTest2, *ResponseTest2](context.Background(), &RequestTest2{Data: "test"})
	assert.Containsf(t, err.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err)
}

func Test_Send_Should_Return_Response_If_Handler_Returns_Success(t *testing.T) {
	handler := &RequestTestHandler{}
	_ = RegisterRequestHandler[*RequestTest, *ResponseTest](handler)
	response, err := Send[*RequestTest, *ResponseTest](context.Background(), &RequestTest{Data: "test"})
	assert.Nil(t, err)
	assert.IsType(t, &ResponseTest{}, response)
	assert.Equal(t, "test", response.Data)
}

///////////////////////////////////////////////////////////////////////////////////////////////
type RequestTest struct {
	Data string
}

type ResponseTest struct {
	Data string
}

type RequestTestHandler struct {
}

func (c *RequestTestHandler) Handle(ctx context.Context, request *RequestTest) (*ResponseTest, error) {
	return &ResponseTest{Data: request.Data}, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
type RequestTest2 struct {
	Data string
}

type ResponseTest2 struct {
	Data string
}

type RequestTestHandler2 struct {
}

func (c *RequestTestHandler2) Handle(ctx context.Context, request *RequestTest2) (*ResponseTest2, error) {
	return &ResponseTest2{Data: request.Data}, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
type RequestTestHandler3 struct {
}

func (c *RequestTestHandler3) Handle(ctx context.Context, request *RequestTest2) (*ResponseTest2, error) {
	return nil, errors.New("some error")
}
