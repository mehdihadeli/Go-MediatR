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
	handle(ctx context.Context, request any, next requestHandlerFunc) (any, IError)
}

type iRequestHandler[TRequest any, TResponse any] interface {
	handle(ctx context.Context, request TRequest) (TResponse, IError)
}

type iNotificationHandler[TNotification any] interface {
	handle(ctx context.Context, notification TNotification) IError
}

var requestHandlersRegistrations = map[reflect.Type]any{}
var notificationHandlersRegistrations = map[reflect.Type][]any{}
var pipelineBehaviours []any

// RegisterRequestHandler register the request handler to mediator registry.
func RegisterRequestHandler[TRequest any, TResponse any](handler iRequestHandler[TRequest, TResponse]) IError {
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

// RegisterNotificationHandler register the notification handler to mediator registry.
func RegisterNotificationHandler[TEvent any](handler iNotificationHandler[TEvent]) {
	var event TEvent
	eventType := reflect.TypeOf(event)

	handlers, exist := notificationHandlersRegistrations[eventType]
	if !exist {
		notificationHandlersRegistrations[eventType] = []any{handler}
	}

	notificationHandlersRegistrations[eventType] = append(handlers, handler)
}

// RegisterNotificationHandlers register the notification handlers to mediator registry.
func RegisterNotificationHandlers[TEvent any](handlers ...iNotificationHandler[TEvent]) {
	for _, handler := range handlers {
		RegisterNotificationHandler[TEvent](handler)
	}
}

func ClearRequestRegistrations() {
	requestHandlersRegistrations = map[reflect.Type]any{}
}

func ClearNotificationRegistrations() {
	notificationHandlersRegistrations = map[reflect.Type][]any{}
}

// Send the request to its corresponding request handler.
func Send[TRequest any, TResponse any](ctx context.Context, request TRequest) (TResponse, IError) {
	requestType := reflect.TypeOf(request)
	var response TResponse
	handler, ok := requestHandlersRegistrations[requestType]
	if !ok {
		// request-response strategy should have exactly one handler and if we can't find a corresponding handler, we should return an IError
		return *new(TResponse), ErrorRequestHandlerNotFound
	}

	requestHandler, ok := handler.(iRequestHandler[TRequest, TResponse])
	if !ok {
		return *new(TResponse), ErrorRequestHandlerNotValid
	}

	if len(pipelineBehaviours) > 0 {
		var reversPipes = reversOrder(pipelineBehaviours)

		var lastHandler requestHandlerFunc = func() (any, IError) {
			return requestHandler.handle(ctx, request)
		}

		aggregateResult := linq.From(reversPipes).AggregateWithSeedT(lastHandler, func(next requestHandlerFunc, pipe iPipelineBehavior) requestHandlerFunc {
			pipeValue := pipe
			nexValue := next

			var handlerFunc requestHandlerFunc = func() (any, IError) {
				return pipeValue.handle(ctx, request, nexValue)
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
		res, err := requestHandler.handle(ctx, request)
		if err != nil {
			// error handling request
			return *new(TResponse), err
		}

		response = res
	}

	return response, nil
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
		notificationHandler, ok := handler.(iNotificationHandler[TNotification])
		if !ok {
			return ErrorNotificationHandlerNotValid
		}
		if err := notificationHandler.handle(ctx, notification); err != nil {
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
