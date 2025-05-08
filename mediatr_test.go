package mediatr

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testData []string
var testMutex sync.Mutex

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

// Helper functions for tests
func countRequestHandlers() int {
	count := 0
	requestHandlersRegistrations.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func countNotificationHandlers(eventType reflect.Type) int {
	if handlers, ok := notificationHandlersRegistrations.Load(eventType); ok {
		return len(handlers.([]interface{}))
	}
	return 0
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

	expectedErr := fmt.Sprintf("handler already exists for type %s", "*mediatr.RequestTest")

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
	assert.Equal(t, 1, countRequestHandlers())
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

	require.NoError(t, err1, "should register first factory without error")
	require.NoError(t, err2, "should register second factory without error")
	assert.Equal(t, 2, countRequestHandlers(), "should have exactly 2 request handlers registered")
}

// Each request should have exactly one handler
func (t *MediatRTests) Test_RegisterRequestHandler_Should_Return_Error_If_Handler_Already_Registered_For_Request() {
	defer cleanup()

	// Expected error message from the package
	expectedErr := fmt.Sprintf("handler already exists for type %s", "*mediatr.RequestTest")

	// Create two handlers for the same request type
	handler1 := &RequestTestHandler{}
	handler2 := &RequestTestHandler{}

	// First registration should succeed
	err1 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler1)
	require.NoError(t, err1, "first registration should succeed")

	// Second registration should fail
	err2 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler2)
	require.Error(t, err2, "second registration should fail")
	assert.Contains(t, err2.Error(), expectedErr,
		"error message should indicate duplicate registration")

	// Verify only one handler is registered
	assert.Equal(t, 1, countRequestHandlers(),
		"only one handler should be registered despite two attempts")
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

	assert.Equal(t, 2, countRequestHandlers())
}

func (t *MediatRTests) Test_Send_Should_Throw_Error_If_No_Handler_Registered() {
	defer cleanup()
	expectedErr := fmt.Sprintf("no handler for request %T", &RequestTest{})
	_, err := Send[*RequestTest, *ResponseTest](context.Background(), &RequestTest{Data: "test"})
	assert.Containsf(t, err.Error(), expectedErr, "expected error containing %q, got %s", expectedErr, err)
}

func (t *MediatRTests) Test_Send_Should_Return_Error_If_Handler_Returns_Error() {
	defer cleanup()

	expectedErr := "some error"

	handler := &RequestTestHandler3{}
	errRegister := RegisterRequestHandler[*RequestTest2, *ResponseTest2](handler)
	require.NoError(t, errRegister, "handler registration should succeed")

	_, err := Send[*RequestTest2, *ResponseTest2](context.Background(), &RequestTest2{Data: "test"})

	require.Error(t, err, "should return error from handler")
	assert.Contains(t, err.Error(), expectedErr,
		"error should contain the original handler error message")
	assert.Contains(t, err.Error(), "handler error:",
		"error should be wrapped with package prefix")
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
	assert.Nil(t, err)

	handler := &RequestTestHandler{}
	errRegister := RegisterRequestHandler[*RequestTest, *ResponseTest](handler)
	assert.Nil(t, errRegister)

	response, err := Send[*RequestTest, *ResponseTest](context.Background(), &RequestTest{Data: "test"})
	assert.Nil(t, err)
	assert.IsType(t, &ResponseTest{}, response)
	assert.Equal(t, "test", response.Data)

	testMutex.Lock()
	defer testMutex.Unlock()
	assert.Contains(t, testData, "PipelineBehaviourTest")
	assert.Contains(t, testData, "PipelineBehaviourTest2")
}

func (t *MediatRTests) Test_RegisterNotificationHandler_Should_Register_Multiple_Handler_For_Notification() {
	defer cleanup()
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler{}

	// Register handlers
	err1 := RegisterNotificationHandler[*NotificationTest](handler1)
	err2 := RegisterNotificationHandler[*NotificationTest](handler2)

	// Verify no errors occurred
	require.NoError(t, err1, "should register first handler without error")
	require.NoError(t, err2, "should register second handler without error")

	// Verify handler count using helper
	notificationType := reflect.TypeOf(&NotificationTest{})
	assert.Equal(t, 2, countNotificationHandlers(notificationType),
		"should have exactly 2 handlers registered for this notification type")
}

func (t *MediatRTests) Test_RegisterNotificationHandlers_Should_Register_Multiple_Handler_For_Notification() {
	defer cleanup()
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler{}
	handler3 := &NotificationTestHandler4{}

	// Register multiple handlers at once
	err := RegisterNotificationHandlers[*NotificationTest](handler1, handler2, handler3)

	// Verify registration was successful
	require.NoError(t, err, "should register multiple handlers without error")

	// Verify handler count using helper
	notificationType := reflect.TypeOf(&NotificationTest{})
	assert.Equal(t, 3, countNotificationHandlers(notificationType),
		"should have exactly 3 handlers registered for this notification type")
}

// notifications could have zero or more handlers
func (t *MediatRTests) Test_Publish_Should_Pass_If_No_Handler_Registered() {
	defer cleanup()
	err := Publish[*NotificationTest](context.Background(), &NotificationTest{})
	assert.Nil(t, err)
}

func (t *MediatRTests) Test_Publish_Should_Return_Error_If_Handler_Returns_Error() {
	defer cleanup()
	expectedErr := "notification handler failed"
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

	count := len(pipelineBehaviors)
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

	assert.Contains(t, err.Error(), "behavior already registered")
}

func (t *MediatRTests) Test_Clear_Request_Registrations() {
	// Setup
	handler1 := &RequestTestHandler{}
	handler2 := &RequestTestHandler2{}

	// Register handlers
	err1 := RegisterRequestHandler[*RequestTest, *ResponseTest](handler1)
	err2 := RegisterRequestHandler[*RequestTest2, *ResponseTest2](handler2)
	require.NoError(t, err1, "should register first handler without error")
	require.NoError(t, err2, "should register second handler without error")

	// Verify initial registration
	assert.Equal(t, 2, countRequestHandlers(), "should have 2 handlers registered before cleanup")

	// Execute
	cleanup()

	// Verify
	assert.Equal(t, 0, countRequestHandlers(), "should have no handlers after cleanup")
}

func (t *MediatRTests) Test_Clear_Notifications_Registrations() {
	// Setup
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler4{}
	notificationType := reflect.TypeOf(&NotificationTest{})

	// Register handlers
	err := RegisterNotificationHandlers[*NotificationTest](handler1, handler2)
	require.NoError(t, err, "should register notification handlers without error")

	// Verify initial registration
	assert.Equal(t, 2, countNotificationHandlers(notificationType),
		"should have 2 notification handlers before clear")

	// Execute
	ClearNotificationRegistrations()

	// Verify
	assert.Equal(t, 0, countNotificationHandlers(notificationType),
		"should have no notification handlers after clear")
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

	res, err := next(ctx)
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

	res, err := next(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// /////////////////////////////////////////////////////////////////////////////////////////////
func cleanup() {
	testMutex.Lock()
	defer testMutex.Unlock()

	// Clear all registrations using package functions
	ClearRequestRegistrations()
	ClearNotificationRegistrations()
	ClearPipelineBehaviors()

	// Reset test data
	testData = nil
}
