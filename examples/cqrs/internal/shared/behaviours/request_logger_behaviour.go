package behaviours

import (
	"context"
	"log"

	"github.com/mehdihadeli/go-mediatr"
)

type RequestLoggerBehaviour struct {
}

func (r *RequestLoggerBehaviour) Handle(ctx context.Context, request interface{}, next mediatr.RequestHandlerFunc) (interface{}, error) {
	log.Printf("logging some stuff before handling the request")

	// https://golang.org/pkg/context/#Context
	ctx = context.WithValue(ctx, "logger_pipeline", true)

	response, err := next(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("logging some stuff after handling the request")

	return response, nil
}
