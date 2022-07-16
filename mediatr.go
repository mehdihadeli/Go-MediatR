package mediatr

import (
	"context"
	"fmt"
	"reflect"
)

type RequestHandler[TRequest any, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, error)
}

var registrations = map[reflect.Type]interface{}{}

type Unit struct{}

// RegisterHandler register the handler to mediatr registry.
func RegisterHandler[TRequest any, TResponse any](h RequestHandler[TRequest, TResponse]) error {
	var request TRequest
	requestType := reflect.TypeOf(request)

	_, exist := registrations[requestType]
	if exist {
		return fmt.Errorf("registerd handler already registered for message %T", requestType)
	}

	registrations[requestType] = h

	return nil
}

// RegisterBehavior TODO
func RegisterBehavior(b interface{}) error {
	return nil
}

// Send the request to its corresponding handler.
func Send[TResponse any, TRequest any](ctx context.Context, request TRequest) (TResponse, error) {

	requestType := reflect.TypeOf(request)

	handler, ok := registrations[requestType]
	if !ok {
		return *new(TResponse), fmt.Errorf("no handlers for command %T", request)
	}

	handlerValue, ok := handler.(RequestHandler[TRequest, TResponse])
	if !ok {
		return *new(TResponse), fmt.Errorf("handler for command %T is not a Handler", request)
	}

	response, err := handlerValue.Handle(ctx, request)
	if err != nil {
		return *new(TResponse), err
	}

	return response, nil
}
