package mediatr

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

type RequestHandler[TRequest any, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, error)
}

type NotificationHandler[TNotification any] interface {
	Handle(ctx context.Context, notification TNotification) error
}

var requestHandlersRegistrations = map[reflect.Type]interface{}{}
var notificationHandlersRegistrations = map[reflect.Type][]interface{}{}

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

// RegisterRequestBehavior TODO
func RegisterRequestBehavior(b interface{}) error {
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

	handler, ok := requestHandlersRegistrations[requestType]
	if !ok {
		return *new(TResponse), errors.Errorf("no handlers for command %T", request)
	}

	handlerValue, ok := handler.(RequestHandler[TRequest, TResponse])
	if !ok {
		return *new(TResponse), errors.Errorf("handler for command %T is not a Handler", request)
	}

	response, err := handlerValue.Handle(ctx, request)
	if err != nil {
		return *new(TResponse), errors.Wrap(err, "error handling request")
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
