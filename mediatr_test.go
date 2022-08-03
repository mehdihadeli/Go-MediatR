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
	errRegister := RegisterRequestHandler[*RequestTest2, *ResponseTest2](handler3)
	if errRegister != nil {
		t.Error(errRegister)
	}
	_, err := Send[*RequestTest2, *ResponseTest2](context.Background(), &RequestTest2{Data: "test"})
	assert.Containsf(t, err.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err)
}

func Test_Send_Should_Dispatch_Request_To_Handler_And_Get_Response(t *testing.T) {
	handler := &RequestTestHandler{}
	errRegister := RegisterRequestHandler[*RequestTest, *ResponseTest](handler)
	if errRegister != nil {
		t.Error(errRegister)
	}
	response, err := Send[*RequestTest, *ResponseTest](context.Background(), &RequestTest{Data: "test"})
	assert.Nil(t, err)
	assert.IsType(t, &ResponseTest{}, response)
	assert.Equal(t, "test", response.Data)
}

func Test_RegisterNotificationHandler_Should_Register_Multiple_Handler_For_Notification(t *testing.T) {
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler{}
	err1 := RegisterNotificationHandler[*NotificationTest](handler1)
	err2 := RegisterNotificationHandler[*NotificationTest](handler2)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func Test_RegisterNotificationHandlers_Should_Register_Multiple_Handler_For_Notification(t *testing.T) {
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler{}
	handler3 := &NotificationTestHandler4{}
	err := RegisterNotificationHandlers[*NotificationTest](handler1, handler2, handler3)

	assert.Nil(t, err)
}

func Test_Publish_Should_Throw_Error_If_No_Handler_Registered(t *testing.T) {
	expectedErr := fmt.Sprintf("no handlers for notification %T", &NotificationTest{})
	err := Publish[*NotificationTest](context.Background(), &NotificationTest{})
	assert.Containsf(t, err.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err)
}

func Test_Publish_Should_Return_Error_If_Handler_Returns_Error(t *testing.T) {
	expectedErr := "error handling notification"
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler{}
	handler3 := &NotificationTestHandler3{}

	errRegister := RegisterNotificationHandlers[*NotificationTest](handler1, handler2, handler3)
	if errRegister != nil {
		t.Error(errRegister)
	}
	err := Publish[*NotificationTest](context.Background(), &NotificationTest{})
	assert.Containsf(t, err.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err)
}

func Test_Publish_Should_Dispatch_Notification_To_All_Handlers_Without_Any_Response_And_Error(t *testing.T) {
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler4{}
	errRegister := RegisterNotificationHandlers[*NotificationTest](handler1, handler2)
	if errRegister != nil {
		t.Error(errRegister)
	}
	err := Publish[*NotificationTest](context.Background(), &NotificationTest{})
	assert.Nil(t, err)
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

///////////////////////////////////////////////////////////////////////////////////////////////
type NotificationTest struct {
	Data string
}

type NotificationTestHandler struct {
}

func (c *NotificationTestHandler) Handle(ctx context.Context, notification *NotificationTest) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
type NotificationTest2 struct {
	Data string
}

type NotificationTestHandler2 struct {
}

func (c *NotificationTestHandler2) Handle(ctx context.Context, notification *NotificationTest2) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////

type NotificationTestHandler3 struct {
}

func (c *NotificationTestHandler3) Handle(ctx context.Context, notification *NotificationTest) error {
	return errors.New("some error")
}

///////////////////////////////////////////////////////////////////////////////////////////////
type NotificationTestHandler4 struct {
}

func (c *NotificationTestHandler4) Handle(ctx context.Context, notification *NotificationTest) error {
	return nil
}
