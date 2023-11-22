package events

import (
	"context"
)

type ProductCreatedEventHandler struct {
}

func NewProductCreatedEventHandler() *ProductCreatedEventHandler {
	return &ProductCreatedEventHandler{}
}

func (c *ProductCreatedEventHandler) Handle(ctx context.Context, event *ProductCreatedEvent) error {
	//Do something with the event here !

	return nil
}
