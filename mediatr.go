//	Mediatr package implements the mediator pattern for Go, providing:
//
// - Request/response message handling
// - Notification broadcasting
// - Pipeline behaviors for cross-cutting concerns
//
// The package is thread-safe and designed for high-performance concurrent use.
package mediatr

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
)

// RequestHandlerFunc is a continuation function used in pipeline behaviors.
// It represents the next handler in the pipeline chain.
type RequestHandlerFunc func(ctx context.Context) (interface{}, error)

// PipelineBehavior defines middleware-like components that can intercept requests.
// Implement this interface to add cross-cutting concerns like logging, validation, etc.
type PipelineBehavior interface {
	Handle(ctx context.Context, request interface{}, next RequestHandlerFunc) (interface{}, error)
}

// RequestHandler handles a specific request type and returns a response.
// Implement this interface for your request handlers.
//
// Example:
//
//	type MyHandler struct{}
//	func (h *MyHandler) Handle(ctx context.Context, req MyRequest) (MyResponse, error) {
//	    // handle request
//	}
type RequestHandler[TRequest any, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, error)
}

// RequestHandlerFactory creates new instances of a request handler.
// Useful when handlers need fresh instances per request.
type RequestHandlerFactory[TRequest any, TResponse any] func() RequestHandler[TRequest, TResponse]

// NotificationHandler processes notifications of a specific type.
// Multiple handlers can process the same notification.
type NotificationHandler[TNotification any] interface {
	Handle(ctx context.Context, notification TNotification) error
}

// NotificationHandlerFactory creates new instances of notification handlers.
type NotificationHandlerFactory[TNotification any] func() NotificationHandler[TNotification]

var (
	requestHandlersRegistrations      sync.Map // map[reflect.Type]interface{}
	notificationHandlersRegistrations sync.Map // map[reflect.Type][]interface{}
	pipelineBehaviors                 []PipelineBehavior

	notificationHandlerMutex sync.Mutex
	pipelineMutex            sync.RWMutex
)

// Unit represents a void return type, used for handlers that don't return data.
type Unit struct{}

// RegisterRequestHandler registers a request handler for a specific request type.
// Returns an error if a handler is already registered for the request type.
//
// Example:
//
//	err := mediatr.RegisterRequestHandler[*MyRequest, *MyResponse](&MyHandler{})
func RegisterRequestHandler[TRequest any, TResponse any](handler RequestHandler[TRequest, TResponse]) error {
	return registerRequestHandler[TRequest, TResponse](handler)
}

// RegisterRequestHandlerFactory registers a factory that creates request handlers.
// Useful for stateful handlers that need fresh instances per request.
func RegisterRequestHandlerFactory[TRequest any, TResponse any](factory RequestHandlerFactory[TRequest, TResponse]) error {
	return registerRequestHandler[TRequest, TResponse](factory)
}

// RegisterRequestPipelineBehaviors registers middleware behaviors that wrap request handlers.
// Behaviors are executed in registration order (first registered runs first).
// Returns error if any behavior is already registered.
func RegisterRequestPipelineBehaviors(behaviours ...PipelineBehavior) error {
	pipelineMutex.Lock()
	defer pipelineMutex.Unlock()

	for _, behavior := range behaviours {
		behaviorType := reflect.TypeOf(behavior)
		for _, existing := range pipelineBehaviors {
			if reflect.TypeOf(existing) == behaviorType {
				return errors.New("behavior already registered")
			}
		}
		pipelineBehaviors = append(pipelineBehaviors, behavior)
	}

	return nil
}

// RegisterNotificationHandler registers a handler for notifications of specific type.
// Multiple handlers can be registered for the same notification type.
func RegisterNotificationHandler[TEvent any](handler NotificationHandler[TEvent]) error {
	return registerNotificationHandler[TEvent](handler)
}

// RegisterNotificationHandlerFactory registers a factory that creates notification handlers.
func RegisterNotificationHandlerFactory[TEvent any](factory NotificationHandlerFactory[TEvent]) error {
	return registerNotificationHandler[TEvent](factory)
}

// RegisterNotificationHandlers registers multiple handlers for a notification type.
// Returns error if no handlers are provided or registration fails.
func RegisterNotificationHandlers[TEvent any](handlers ...NotificationHandler[TEvent]) error {
	if len(handlers) == 0 {
		return errors.New("no handlers provided")
	}

	for _, handler := range handlers {
		err := RegisterNotificationHandler(handler)
		if err != nil {
			return err
		}
	}

	return nil
}

// RegisterNotificationHandlersFactories registers multiple handler factories.
func RegisterNotificationHandlersFactories[TEvent any](factories ...NotificationHandlerFactory[TEvent]) error {
	if len(factories) == 0 {
		return errors.New("no handlers provided")
	}

	for _, handler := range factories {
		err := RegisterNotificationHandlerFactory[TEvent](handler)
		if err != nil {
			return err
		}
	}

	return nil
}

// Send dispatches a request to its registered handler and returns the response.
// Executes all registered pipeline behaviors in order.
// Returns error if:
// - No handler is registered for the request
// - Handler returns an error
// - Any pipeline behavior returns an error
//
// Example:
//
//	response, err := mediatr.Send[*MyRequest, *MyResponse](ctx, &MyRequest{})
func Send[TRequest any, TResponse any](ctx context.Context, request TRequest) (TResponse, error) {
	requestType := reflect.TypeOf(request)

	handler, ok := requestHandlersRegistrations.Load(requestType)
	if !ok {
		return *new(TResponse), errors.Errorf("no handler for request %T", request)
	}

	pipelineMutex.RLock()
	behaviors := make([]PipelineBehavior, len(pipelineBehaviors))
	copy(behaviors, pipelineBehaviors)
	pipelineMutex.RUnlock()

	handlerValue, ok := buildRequestHandler[TRequest, TResponse](handler)
	if !ok {
		return *new(TResponse), errors.Errorf("invalid handler for request %T", request)
	}

	if len(behaviors) > 0 {
		result, err := buildPipeline(behaviors, handlerValue, request)(ctx)
		if err != nil {
			return *new(TResponse), errors.Wrap(err, "pipeline error")
		}
		return result.(TResponse), nil
	}

	response, err := handlerValue.Handle(ctx, request)
	if err != nil {
		return *new(TResponse), errors.Wrap(err, "handler error")
	}

	return response, nil
}

// Publish broadcasts a notification to all registered handlers.
// All handlers are executed, even if some return errors.
// Returns the first error encountered, if any.
//
// Example:
//
//	type OrderShipped struct { OrderID string }
//
//	// Register handlers
//	mediatr.RegisterNotificationHandlers[OrderShipped](
//	    &ShippingNotifier{},
//	    &InventoryUpdater{},
//	)
//
//	// Publish
//	err := mediatr.Publish(ctx, OrderShipped{OrderID: "123"})
//	if err != nil { /* handle error */ }
func Publish[TNotification any](ctx context.Context, notification TNotification) error {
	eventType := reflect.TypeOf(notification)

	handlers, ok := notificationHandlersRegistrations.Load(eventType)
	if !ok {
		return nil
	}

	handlerList := handlers.([]interface{})

	for _, handler := range handlerList {
		handlerValue, ok := buildNotificationHandler[TNotification](handler)
		if !ok {
			return errors.Errorf("invalid handler type for notification %T", notification)
		}
		if err := handlerValue.Handle(ctx, notification); err != nil {
			return errors.Wrap(err, "notification handler failed")
		}
	}

	return nil
}

// ClearRequestRegistrations removes all registered request handlers.
// Useful for testing scenarios.
func ClearRequestRegistrations() {
	requestHandlersRegistrations = sync.Map{}
}

// ClearNotificationRegistrations removes all registered notification handlers.
func ClearNotificationRegistrations() {
	notificationHandlersRegistrations = sync.Map{}
}

// ClearPipelineBehaviors removes all registered pipeline behaviors.
func ClearPipelineBehaviors() {
	pipelineMutex.Lock()
	defer pipelineMutex.Unlock()
	pipelineBehaviors = []PipelineBehavior{}
}

func registerRequestHandler[TRequest any, TResponse any](handler any) error {
	var request TRequest
	requestType := reflect.TypeOf(request)

	if _, loaded := requestHandlersRegistrations.LoadOrStore(requestType, handler); loaded {
		return errors.Errorf("handler already exists for type %s", requestType.String())
	}
	return nil
}

func registerNotificationHandler[TEvent any](handler any) error {
	var event TEvent
	eventType := reflect.TypeOf(event)

	// Uses separate mutex for slice modifications and adding new item with LoadOrStore if not exists for prevention conflict with concurrent goroutines
	notificationHandlerMutex.Lock()
	defer notificationHandlerMutex.Unlock()

	// If not found, stores a new slice with the handler as its first element, If found, returns the existing slice of handlers.
	actual, loaded := notificationHandlersRegistrations.LoadOrStore(eventType, []interface{}{handler})
	if !loaded {
		return nil
	}

	handlers := actual.([]interface{})

	newHandlers := append(handlers, handler)
	notificationHandlersRegistrations.Store(eventType, newHandlers)

	return nil
}

func buildRequestHandler[TRequest any, TResponse any](handler any) (RequestHandler[TRequest, TResponse], bool) {
	handlerValue, ok := handler.(RequestHandler[TRequest, TResponse])
	if !ok {
		factory, ok := handler.(RequestHandlerFactory[TRequest, TResponse])
		if !ok {
			return nil, false
		}

		return factory(), true
	}

	return handlerValue, true
}

func buildNotificationHandler[TNotification any](handler any) (NotificationHandler[TNotification], bool) {
	handlerValue, ok := handler.(NotificationHandler[TNotification])
	if !ok {
		factory, ok := handler.(NotificationHandlerFactory[TNotification])
		if !ok {
			return nil, false
		}

		return factory(), true
	}

	return handlerValue, true
}

// buildPipeline constructs the middleware chain
func buildPipeline[TRequest any, TResponse any](
	behaviors []PipelineBehavior,
	handler RequestHandler[TRequest, TResponse],
	request TRequest,
) RequestHandlerFunc {
	reversed := reverseBehaviors(behaviors)

	chain := func(ctx context.Context) (interface{}, error) {
		return handler.Handle(ctx, request)
	}

	// Build the pipeline by wrapping each behavior
	for _, behavior := range reversed {
		currentBehavior := behavior // capture for closure
		next := chain
		chain = func(ctx context.Context) (interface{}, error) {
			return currentBehavior.Handle(ctx, request, next)
		}
	}

	return chain
}

// reverseBehaviors reverses the order of pipeline behaviors
func reverseBehaviors(behaviors []PipelineBehavior) []PipelineBehavior {
	reversed := make([]PipelineBehavior, len(behaviors))
	for i := range behaviors {
		reversed[len(behaviors)-1-i] = behaviors[i]
	}
	return reversed
}
