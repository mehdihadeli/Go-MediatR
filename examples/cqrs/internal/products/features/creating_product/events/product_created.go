package events

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type ProductCreatedEvent struct {
	ProductID   uuid.UUID `json:"product_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewProductCreatedEvent(id uuid.UUID, name string, description string, price float64, createdAt time.Time) *ProductCreatedEvent {
	return &ProductCreatedEvent{ProductID: id, Name: name, Description: description, Price: price, CreatedAt: createdAt}
}
