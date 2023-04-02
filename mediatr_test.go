package mediatr

import (
	"context"
	"fmt"
	"testing"

	"github.com/goccy/go-reflect"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testData []string

func TestRunner(t *testing.T) {
	//https://pkg.go.dev/testing@master#hdr-Subtests_and_Sub_benchmarks
	t.Run("A=request-response", func(t *testing.T) {
		test := MediatRTests{T: t}
		test.Test_RegisterRequestHandler_Should_Return_Error_If_Handler_Already_Registered_For_Request()
		test.Test_RegisterRequestHandler_Should_Register_All_Handlers_For_Different_Requests()
		test.Test_Send_Should_Throw_Error_If_No_Handler_Registered()
		test.Test_Send_Should_Return_Error_If_Handler_Returns_Error()
		test.Test_Send_Should_Dispatch_Request_To_Handler_And_Get_Response_Without_Pipeline()
		test.Test_Clear_Request_Registrations()

		test.Test_RegisterRequestHandlerFactory_Should_Return_Error_If_Handler_Already_Registered_For_Request()
		test.Test_RegisterRequestHandlerFactory_Should_Register_All_Handlers_For_Different_Requests()
		test.Test_Send_Should_Dispatch_Request_To_Factory()
	})

	t.Run("B=notifications", func(t *testing.T) {
		test := MediatRTests{T: t}
		test.Test_Publish_Should_Pass_If_No_Handler_Registered()
		test.Test_Publish_Should_Return_Error_If_Handler_Returns_Error()
		test.Test_Publish_Should_Dispatch_Notification_To_All_Handlers_Without_Any_Response_And_Error()
		test.Test_Clear_Notifications_Registrations()

		test.Test_Publish_Should_Dispatch_Notification_To_All_Handlers_Factories_Without_Any_Response_And_Error()
	})

	t.Run("C=pipeline-behaviours", func(t *testing.T) {
		test := MediatRTests{T: t}
		test.Test_Register_Behaviours_Should_Register_Behaviours_In_The_Registry_Correctly()
		test.Test_Register_Duplicate_Behaviours_Should_Throw_Error()
		test.Test_Send_Should_Dispatch_Request_To_Handler_And_Get_Response_With_Pipeline()
	})
}

type MediatRTests struct {
	*testing.T
}

func (t *MediatRTests) Test_Send_Should_Dispatch_Request_To_Factory() {
	defer cleanup()
	var factory1 RequestHandlerFactory[*RequestTest, *ResponseTest] = func() RequestHandler[*RequestTest, *ResponseTest] {
		return &RequestTestHandler{}
	}
	errRegister := RegisterRequestHandlerFactory(factory1)
	if errRegister != nil {
		t.Error(errRegister)
	}

	response, err := Send[*RequestTest, *ResponseTest](context.Background(), &RequestTest{Data: "test"})
	assert.Nil(t, err)
	assert.IsType(t, &ResponseTest{}, response)
	assert.Equal(t, "test", response.Data)
}

func (t *MediatRTests) Test_RegisterRequestHandlerFactory_Should_Return_Error_If_Handler_Already_Registered_For_Request() {
	defer cleanup()

	expectedErr := fmt.Sprintf("registered handler already exists in the registry for message %s", "*mediatr.RequestTest")

	var factory1 RequestHandlerFactory[*RequestTest, *ResponseTest] = func() RequestHandler[*RequestTest, *ResponseTest] {
		return &RequestTestHandler{}
	}
	var factory2 RequestHandlerFactory[*RequestTest, *ResponseTest] = func() RequestHandler[*RequestTest, *ResponseTest] {
		return &RequestTestHandler{}
	}

	err1 := RegisterRequestHandlerFactory(factory1)
	err2 := RegisterRequestHandlerFactory(factory2)

	assert.Nil(t, err1)
	assert.Containsf(t, err2.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err2)

	count := len(requestHandlersRegistrations)
	assert.Equal(t, 1, count)
}

func (t *MediatRTests) Test_RegisterRequestHandlerFactory_Should_Register_All_Handlers_For_Different_Requests() {
	defer cleanup()
	var factory1 RequestHandlerFactory[*RequestTest, *ResponseTest] = func() RequestHandler[*RequestTest, *ResponseTest] {
		return &RequestTestHandler{}
	}
	var factory2 RequestHandlerFactory[*RequestTest2, *ResponseTest2] = func() RequestHandler[*RequestTest2, *ResponseTest2] {
		return &RequestTestHandler2{}
	}

	err1 := RegisterRequestHandlerFactory(factory1)
	err2 := RegisterRequestHandlerFactory(factory2)

	if err1 != nil {
		t.Errorf("error registering request handler: %s", err1)
	}

	if err2 != nil {
		t.Errorf("error registering request handler: %s", err2)
	}

	count := len(requestHandlersRegistrations)
	assert.Equal(t, 2, count)
}

// Each request should have exactly one handler
func (t *MediatRTests) Test_RegisterRequestHandler_Should_Return_Error_If_Handler_Already_Registered_For_Request() {
	defer cleanup()
	expectedErr := fmt.Sprintf("registered handler already exists in the registry for message %s", "*mediatr.RequestTest")
	handler1 := &RequestTestHandler{}
	handler2 := &RequestTestHandler{}
	err1 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler1)
	err2 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler2)

	assert.Nil(t, err1)
	assert.Containsf(t, err2.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err2)

	count := len(requestHandlersRegistrations)
	assert.Equal(t, 1, count)
}

func (t *MediatRTests) Test_RegisterRequestHandler_Should_Register_All_Handlers_For_Different_Requests() {
	defer cleanup()
	handler1 := &RequestTestHandler{}
	handler2 := &RequestTestHandler2{}
	err1 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler1)
	err2 := RegisterRequestHandler[*RequestTest2, *ResponseTest2](handler2)

	if err1 != nil {
		t.Errorf("error registering request handler: %s", err1)
	}

	if err2 != nil {
		t.Errorf("error registering request handler: %s", err2)
	}

	count := len(requestHandlersRegistrations)
	assert.Equal(t, 2, count)
}

func (t *MediatRTests) Test_Send_Should_Throw_Error_If_No_Handler_Registered() {
	defer cleanup()
	expectedErr := fmt.Sprintf("no handler for request %T", &RequestTest{})
	_, err := Send[*RequestTest, *ResponseTest](context.Background(), &RequestTest{Data: "test"})
	assert.Containsf(t, err.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err)
}

func (t *MediatRTests) Test_Send_Should_Return_Error_If_Handler_Returns_Error() {
	defer cleanup()
	expectedErr := "error handling request"
	handler3 := &RequestTestHandler3{}
	errRegister := RegisterRequestHandler[*RequestTest2, *ResponseTest2](handler3)
	if errRegister != nil {
		t.Error(errRegister)
	}
	_, err := Send[*RequestTest2, *ResponseTest2](context.Background(), &RequestTest2{Data: "test"})
	assert.Containsf(t, err.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err)
}

func (t *MediatRTests) Test_Send_Should_Dispatch_Request_To_Handler_And_Get_Response_Without_Pipeline() {
	defer cleanup()
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

func (t *MediatRTests) Test_Send_Should_Dispatch_Request_To_Handler_And_Get_Response_With_Pipeline() {
	defer cleanup()
	pip1 := &PipelineBehaviourTest{}
	pip2 := &PipelineBehaviourTest2{}
	err := RegisterRequestPipelineBehaviors(pip1, pip2)
	if err != nil {
		t.Errorf("error registering request pipeline behaviors: %s", err)
	}

	handler := &RequestTestHandler{}
	errRegister := RegisterRequestHandler[*RequestTest, *ResponseTest](handler)
	if errRegister != nil {
		t.Error(errRegister)
	}

	response, err := Send[*RequestTest, *ResponseTest](context.Background(), &RequestTest{Data: "test"})
	assert.Nil(t, err)
	assert.IsType(t, &ResponseTest{}, response)
	assert.Equal(t, "test", response.Data)
	assert.Contains(t, testData, "PipelineBehaviourTest")
	assert.Contains(t, testData, "PipelineBehaviourTest2")
}

func (t *MediatRTests) Test_RegisterNotificationHandler_Should_Register_Multiple_Handler_For_Notification() {
	defer cleanup()
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler{}
	err1 := RegisterNotificationHandler[*NotificationTest](handler1)
	err2 := RegisterNotificationHandler[*NotificationTest](handler2)

	if err1 != nil {
		t.Errorf("error registering notification handler: %s", err1)
	}
	if err2 != nil {
		t.Errorf("error registering notification handler: %s", err2)
	}

	count := len(notificationHandlersRegistrations[reflect.TypeOf(&NotificationTest{})])
	assert.Equal(t, 2, count)
}

func (t *MediatRTests) Test_RegisterNotificationHandlers_Should_Register_Multiple_Handler_For_Notification() {
	defer cleanup()
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler{}
	handler3 := &NotificationTestHandler4{}
	err := RegisterNotificationHandlers[*NotificationTest](handler1, handler2, handler3)

	if err != nil {
		t.Errorf("error registering notification handlers: %s", err)
	}

	count := len(notificationHandlersRegistrations[reflect.TypeOf(&NotificationTest{})])
	assert.Equal(t, 3, count)
}

// notifications could have zero or more handlers
func (t *MediatRTests) Test_Publish_Should_Pass_If_No_Handler_Registered() {
	defer cleanup()
	err := Publish[*NotificationTest](context.Background(), &NotificationTest{})
	assert.Nil(t, err)
}

func (t *MediatRTests) Test_Publish_Should_Return_Error_If_Handler_Returns_Error() {
	defer cleanup()
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

func (t *MediatRTests) Test_Publish_Should_Dispatch_Notification_To_All_Handlers_Factories_Without_Any_Response_And_Error() {
	defer cleanup()
	var factory1 NotificationHandlerFactory[*NotificationTest] = func() NotificationHandler[*NotificationTest] {
		return &NotificationTestHandler{}
	}
	var factory2 NotificationHandlerFactory[*NotificationTest] = func() NotificationHandler[*NotificationTest] {
		return &NotificationTestHandler4{}
	}

	errRegister := RegisterNotificationHandlersFactories(factory1, factory2)
	if errRegister != nil {
		t.Error(errRegister)
	}

	notification := &NotificationTest{}
	err := Publish[*NotificationTest](context.Background(), notification)
	assert.Nil(t, err)
	assert.True(t, notification.Processed)
}

func (t *MediatRTests) Test_Publish_Should_Dispatch_Notification_To_All_Handlers_Without_Any_Response_And_Error() {
	defer cleanup()
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler4{}
	errRegister := RegisterNotificationHandlers[*NotificationTest](handler1, handler2)
	if errRegister != nil {
		t.Error(errRegister)
	}

	notification := &NotificationTest{}
	err := Publish[*NotificationTest](context.Background(), notification)
	assert.Nil(t, err)
	assert.True(t, notification.Processed)
}

func (t *MediatRTests) Test_Register_Behaviours_Should_Register_Behaviours_In_The_Registry_Correctly() {
	defer cleanup()
	pip1 := &PipelineBehaviourTest{}
	pip2 := &PipelineBehaviourTest2{}

	err := RegisterRequestPipelineBehaviors(pip1, pip2)
	if err != nil {
		t.Errorf("error registering behaviours: %s", err)
	}

	count := len(pipelineBehaviours)
	assert.Equal(t, 2, count)
}

func (t *MediatRTests) Test_Register_Duplicate_Behaviours_Should_Throw_Error() {
	defer cleanup()
	pip1 := &PipelineBehaviourTest{}
	pip2 := &PipelineBehaviourTest{}
	err := RegisterRequestPipelineBehaviors(pip1, pip2)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	assert.Contains(t, err.Error(), "registered behavior already exists in the registry")
}

func (t *MediatRTests) Test_Clear_Request_Registrations() {
	handler1 := &RequestTestHandler{}
	handler2 := &RequestTestHandler2{}
	err1 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler1)
	err2 := RegisterRequestHandler[*RequestTest2, *ResponseTest2](handler2)
	require.NoError(t, err1, err2)

	ClearRequestRegistrations()

	count := len(requestHandlersRegistrations)
	assert.Equal(t, 0, count)
}

func (t *MediatRTests) Test_Clear_Notifications_Registrations() {
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler4{}
	errRegister := RegisterNotificationHandlers[*NotificationTest](handler1, handler2)
	require.NoError(t, errRegister)

	ClearNotificationRegistrations()

	count := len(notificationHandlersRegistrations)
	assert.Equal(t, 0, count)
}

// /////////////////////////////////////////////////////////////////////////////////////////////
type RequestTest struct {
	Data string
}

type ResponseTest struct {
	Data string
}

type RequestTestHandler struct {
}

func (c *RequestTestHandler) Handle(ctx context.Context, request *RequestTest) (*ResponseTest, error) {
	fmt.Println("RequestTestHandler.Handled")
	testData = append(testData, "RequestTestHandler")

	return &ResponseTest{Data: request.Data}, nil
}

// /////////////////////////////////////////////////////////////////////////////////////////////
type RequestTest2 struct {
	Data string
}

type ResponseTest2 struct {
	Data string
}

type RequestTestHandler2 struct {
}

func (c *RequestTestHandler2) Handle(ctx context.Context, request *RequestTest2) (*ResponseTest2, error) {
	fmt.Println("RequestTestHandler2.Handled")
	testData = append(testData, "RequestTestHandler2")

	return &ResponseTest2{Data: request.Data}, nil
}

// /////////////////////////////////////////////////////////////////////////////////////////////
type RequestTestHandler3 struct {
}

func (c *RequestTestHandler3) Handle(ctx context.Context, request *RequestTest2) (*ResponseTest2, error) {
	return nil, errors.New("some error")
}

// /////////////////////////////////////////////////////////////////////////////////////////////
type NotificationTest struct {
	Data      string
	Processed bool
}

type NotificationTestHandler struct {
}

func (c *NotificationTestHandler) Handle(ctx context.Context, notification *NotificationTest) error {
	notification.Processed = true
	fmt.Println("NotificationTestHandler.Handled")
	testData = append(testData, "NotificationTestHandler")

	return nil
}

// /////////////////////////////////////////////////////////////////////////////////////////////
type NotificationTest2 struct {
	Data      string
	Processed bool
}

type NotificationTestHandler2 struct {
}

func (c *NotificationTestHandler2) Handle(ctx context.Context, notification *NotificationTest2) error {
	notification.Processed = true
	fmt.Println("NotificationTestHandler2.Handled")
	testData = append(testData, "NotificationTestHandler2")

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////

type NotificationTestHandler3 struct {
}

func (c *NotificationTestHandler3) Handle(ctx context.Context, notification *NotificationTest) error {
	return errors.New("some error")
}

// /////////////////////////////////////////////////////////////////////////////////////////////
type NotificationTestHandler4 struct {
}

func (c *NotificationTestHandler4) Handle(ctx context.Context, notification *NotificationTest) error {
	notification.Processed = true
	fmt.Println("NotificationTestHandler4.Handled")
	testData = append(testData, "NotificationTestHandler4")

	return nil
}

// /////////////////////////////////////////////////////////////////////////////////////////////
type PipelineBehaviourTest struct {
}

func (c *PipelineBehaviourTest) Handle(ctx context.Context, request interface{}, next RequestHandlerFunc) (interface{}, error) {
	fmt.Println("PipelineBehaviourTest.Handled")
	testData = append(testData, "PipelineBehaviourTest")

	res, err := next()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// /////////////////////////////////////////////////////////////////////////////////////////////
type PipelineBehaviourTest2 struct {
}

func (c *PipelineBehaviourTest2) Handle(ctx context.Context, request interface{}, next RequestHandlerFunc) (interface{}, error) {
	fmt.Println("PipelineBehaviourTest2.Handled")
	testData = append(testData, "PipelineBehaviourTest2")

	res, err := next()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// /////////////////////////////////////////////////////////////////////////////////////////////
func cleanup() {
	requestHandlersRegistrations = map[reflect.Type]interface{}{}
	notificationHandlersRegistrations = map[reflect.Type][]interface{}{}
	pipelineBehaviours = []interface{}{}
}
