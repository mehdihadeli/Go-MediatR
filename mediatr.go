package mediatr

import (
	"context"
	"github.com/ahmetb/go-linq/v3"
	"github.com/goccy/go-reflect"
	"github.com/pkg/errors"
)

type RequestHandlerFunc func() (interface{}, error)

type RequestHandler[TRequest any, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, error)
}

type NotificationHandler[TNotification any] interface {
	Handle(ctx context.Context, notification TNotification) error
}

type PipelineBehavior interface {
	Handle(ctx context.Context, request interface{}, next RequestHandlerFunc) (interface{}, error)
}

var requestHandlersRegistrations = map[reflect.Type]interface{}{}
var notificationHandlersRegistrations = map[reflect.Type][]interface{}{}
var pipelineBehaviours []interface{} = []interface{}{}

type Unit struct{}

// RegisterRequestHandler register the request handler to mediatr registry.
func RegisterRequestHandler[TRequest any, TResponse any](handler RequestHandler[TRequest, TResponse]) error {
	var request TRequest
	requestType := reflect.TypeOf(request)

	_, exist := requestHandlersRegistrations[requestType]
	if exist {
		// each request in request/response strategy should have just one handler
		return errors.Errorf("registered handler already exists in the registry for message %s", requestType.String())
	}

	requestHandlersRegistrations[requestType] = handler

	return nil
}

// RegisterRequestPipelineBehaviors register the request behaviors to mediatr registry.
func RegisterRequestPipelineBehaviors(behaviours ...PipelineBehavior) error {
	for _, behavior := range behaviours {
		behaviorType := reflect.TypeOf(behavior)

		existsPipe := existsPipeType(behaviorType)
		if existsPipe {
			return errors.Errorf("registered behavior already exists in the registry.")
		}

		pipelineBehaviours = append(pipelineBehaviours, behavior)
	}

	return nil
}

// RegisterNotificationHandler register the notification handler to mediatr registry.
func RegisterNotificationHandler[TEvent any](handler NotificationHandler[TEvent]) error {
	var event TEvent
	eventType := reflect.TypeOf(event)

	handlers, exist := notificationHandlersRegistrations[eventType]
	if exist == false {
		notificationHandlersRegistrations[eventType] = []interface{}{handler}
		return nil
	}

	notificationHandlersRegistrations[eventType] = append(handlers, handler)

	return nil
}

// RegisterNotificationHandlers register the notification handlers to mediatr registry.
func RegisterNotificationHandlers[TEvent any](handlers ...NotificationHandler[TEvent]) error {
	if len(handlers) == 0 {
		return errors.New("no handlers provided")
	}

	for _, handler := range handlers {
		err := RegisterNotificationHandler[TEvent](handler)
		if err != nil {
			return err
		}
	}

	return nil
}

// Send the request to its corresponding request handler.
func Send[TRequest any, TResponse any](ctx context.Context, request TRequest) (TResponse, error) {
	requestType := reflect.TypeOf(request)
	var response TResponse
	handler, ok := requestHandlersRegistrations[requestType]
	if !ok {
		return *new(TResponse), errors.Errorf("no handlers for request %T", request)
	}

	handlerValue, ok := handler.(RequestHandler[TRequest, TResponse])
	if !ok {
		return *new(TResponse), errors.Errorf("handler for request %T is not a Handler", request)
	}

	if len(pipelineBehaviours) > 0 {
		var reversPipes = reversOrder(pipelineBehaviours)

		var lastHandler RequestHandlerFunc = func() (interface{}, error) {
			return handlerValue.Handle(ctx, request)
		}

		aggregateResult := linq.From(reversPipes).AggregateWithSeedT(lastHandler, func(next RequestHandlerFunc, pipe PipelineBehavior) RequestHandlerFunc {
			pipeValue := pipe
			nexValue := next

			var handlerFunc RequestHandlerFunc = func() (interface{}, error) {
				return pipeValue.Handle(ctx, request, nexValue)
			}

			return handlerFunc
		})

		v := aggregateResult.(RequestHandlerFunc)
		response, err := v()

		if err != nil {
			return *new(TResponse), errors.Wrap(err, "error handling request")
		}

		return response.(TResponse), nil
	} else {
		res, err := handlerValue.Handle(ctx, request)
		if err != nil {
			return *new(TResponse), errors.Wrap(err, "error handling request")
		}

		response = res
	}

	return response, nil
}

// Publish the notification event to its corresponding notification handler.
func Publish[TNotification any](ctx context.Context, notification TNotification) error {
	eventType := reflect.TypeOf(notification)

	handlers, ok := notificationHandlersRegistrations[eventType]
	if !ok {
		return errors.Errorf("no handlers for notification %T", notification)
	}

	for _, handler := range handlers {
		handlerValue, ok := handler.(NotificationHandler[TNotification])
		if !ok {
			return errors.Errorf("handler for notification %T is not a Handler", notification)
		}

		err := handlerValue.Handle(ctx, notification)
		if err != nil {
			return errors.Wrap(err, "error handling notification")
		}
	}

	return nil
}

func reversOrder(values []interface{}) []interface{} {
	var reverseValues []interface{}

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
