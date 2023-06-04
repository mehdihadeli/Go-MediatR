package mediator

import (
	"context"
	"github.com/ahmetb/go-linq/v3"
	"github.com/goccy/go-reflect"
)

// requestHandlerFunc is a continuation for the next task to execute in the pipeline
type requestHandlerFunc func() (any, IError)

// iPipelineBehavior is a Pipeline behavior for wrapping the inner handler.
type iPipelineBehavior interface {
	Handle(ctx context.Context, request any, next requestHandlerFunc) (any, IError)
}

type iRequestHandler[TRequest any, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, IError)
}

type RequestHandlerFactory[TRequest any, TResponse any] func() iRequestHandler[TRequest, TResponse]

type iNotificationHandler[TNotification any] interface {
	Handle(ctx context.Context, notification TNotification) IError
}

type NotificationHandlerFactory[TNotification any] func() iNotificationHandler[TNotification]

var requestHandlersRegistrations = map[reflect.Type]any{}
var notificationHandlersRegistrations = map[reflect.Type][]any{}
var pipelineBehaviours []any

func registerRequestHandler[TRequest any, TResponse any](handler any) error {
	var request TRequest
	requestType := reflect.TypeOf(request)

	_, exist := requestHandlersRegistrations[requestType]
	if exist {
		// each request in request/response strategy should have just one handler
		return ErrorRequestHandlerAlreadyExists
	}

	requestHandlersRegistrations[requestType] = handler

	return nil
}

// RegisterRequestHandler register the request handler to mediatr registry.
func RegisterRequestHandler[TRequest any, TResponse any](handler iRequestHandler[TRequest, TResponse]) error {
	return registerRequestHandler[TRequest, TResponse](handler)
}

// RegisterRequestHandlerFactory register the request handler factory to mediatr registry.
func RegisterRequestHandlerFactory[TRequest any, TResponse any](factory RequestHandlerFactory[TRequest, TResponse]) error {
	return registerRequestHandler[TRequest, TResponse](factory)
}

// RegisterRequestPipelineBehaviors register the request behaviors to mediator registry.
func RegisterRequestPipelineBehaviors(behaviours ...iPipelineBehavior) IError {
	for _, behavior := range behaviours {
		behaviorType := reflect.TypeOf(behavior)

		existsPipe := existsPipeType(behaviorType)
		if existsPipe {
			return ErrorRequestPipelineBehaviorAlreadyExists
		}

		pipelineBehaviours = append(pipelineBehaviours, behavior)
	}

	return nil
}

func registerNotificationHandler[TEvent any](handler any) {
	var event TEvent
	eventType := reflect.TypeOf(event)

	handlers, exist := notificationHandlersRegistrations[eventType]
	if !exist {
		notificationHandlersRegistrations[eventType] = []any{handler}
	}

	notificationHandlersRegistrations[eventType] = append(handlers, handler)
}

// RegisterNotificationHandler register the notification handler to mediatr registry.
func RegisterNotificationHandler[TEvent any](handler iNotificationHandler[TEvent]) {
	registerNotificationHandler[TEvent](handler)
}

// RegisterNotificationHandlerFactory register the notification handler factory to mediatr registry.
func RegisterNotificationHandlerFactory[TEvent any](factory NotificationHandlerFactory[TEvent]) {
	registerNotificationHandler[TEvent](factory)
}

// RegisterNotificationHandlers register the notification handlers to mediator registry.
func RegisterNotificationHandlers[TEvent any](handlers ...iNotificationHandler[TEvent]) {
	for _, handler := range handlers {
		RegisterNotificationHandler[TEvent](handler)
	}
}

// RegisterNotificationHandlersFactories register the notification handlers factories to mediatr registry.
func RegisterNotificationHandlersFactories[TEvent any](factories ...NotificationHandlerFactory[TEvent]) {
	for _, handler := range factories {
		RegisterNotificationHandlerFactory[TEvent](handler)
	}
}

func ClearRequestRegistrations() {
	requestHandlersRegistrations = map[reflect.Type]any{}
}

func ClearNotificationRegistrations() {
	notificationHandlersRegistrations = map[reflect.Type][]any{}
}

func buildRequestHandler[TRequest any, TResponse any](handler any) (iRequestHandler[TRequest, TResponse], bool) {
	handlerValue, ok := handler.(iRequestHandler[TRequest, TResponse])
	if !ok {
		factory, ok := handler.(RequestHandlerFactory[TRequest, TResponse])
		if !ok {
			return nil, false
		}

		return factory(), true
	}

	return handlerValue, true
}

// Send the request to its corresponding request handler.
func Send[TRequest any, TResponse any](ctx context.Context, request TRequest) (TResponse, IError) {
	requestType := reflect.TypeOf(request)
	var response TResponse
	handler, ok := requestHandlersRegistrations[requestType]
	if !ok {
		// request-response strategy should have exactly one handler and if we can't find a corresponding handler, we should return an iError
		return *new(TResponse), ErrorRequestHandlerNotFound
	}

	requestHandler, ok := buildRequestHandler[TRequest, TResponse](handler)
	if !ok {
		return *new(TResponse), ErrorRequestHandlerNotValid
	}

	if len(pipelineBehaviours) > 0 {
		var reversPipes = reversOrder(pipelineBehaviours)

		var lastHandler requestHandlerFunc = func() (any, IError) {
			return requestHandler.Handle(ctx, request)
		}

		aggregateResult := linq.From(reversPipes).AggregateWithSeedT(lastHandler, func(next requestHandlerFunc, pipelineBehavior iPipelineBehavior) requestHandlerFunc {
			nexValue := next
			var handlerFunc requestHandlerFunc = func() (any, IError) {
				return pipelineBehavior.Handle(ctx, request, nexValue)
			}
			return handlerFunc
		})

		v := aggregateResult.(requestHandlerFunc)
		response, err := v()

		if err != nil {
			// error handling request
			return *new(TResponse), err
		}

		return response.(TResponse), nil
	} else {
		res, err := requestHandler.Handle(ctx, request)
		if err != nil {
			// error handling request
			return *new(TResponse), err
		}

		response = res
	}

	return response, nil
}

func buildNotificationHandler[TNotification any](handler any) (iNotificationHandler[TNotification], bool) {
	handlerValue, ok := handler.(iNotificationHandler[TNotification])
	if !ok {
		factory, ok := handler.(NotificationHandlerFactory[TNotification])
		if !ok {
			return nil, false
		}

		return factory(), true
	}

	return handlerValue, true
}

// Publish the notification event to its corresponding notification handler.
func Publish[TNotification any](ctx context.Context, notification TNotification) IError {
	eventType := reflect.TypeOf(notification)

	handlers, ok := notificationHandlersRegistrations[eventType]
	if !ok {
		// notification strategy should have zero or more handlers, so it should run without any error if we can't find a corresponding handler
		return nil
	}

	for _, handler := range handlers {
		notificationHandler, ok := buildNotificationHandler[TNotification](handler)
		if !ok {
			return ErrorNotificationHandlerNotValid
		}
		if err := notificationHandler.Handle(ctx, notification); err != nil {
			return err
		}
	}

	return nil
}

func reversOrder(values []any) []any {
	var reverseValues []any

	for i := len(values) - 1; i >= 0; i-- {
		reverseValues = append(reverseValues, values[i])
	}

	return reverseValues
}

func existsPipeType(p reflect.Type) bool {
	for _, pipe := range pipelineBehaviours {
		if reflect.TypeOf(pipe) == p {
			return true
		}
	}

	return false
}
