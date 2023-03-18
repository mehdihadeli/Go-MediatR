package behaviours

import (
	"context"
	"github.com/ehsandavari/go-mediator"
	"log"
)

type RequestLoggerBehaviour struct {
}

func (r *RequestLoggerBehaviour) Handle(ctx context.Context, request interface{}, next mediator.RequestHandlerFunc) (interface{}, error) {
	log.Printf("logging some stuff before handling the request")

	response, err := next()
	if err != nil {
		return nil, err
	}

	log.Println("logging some stuff after handling the request")

	return response, nil
}
