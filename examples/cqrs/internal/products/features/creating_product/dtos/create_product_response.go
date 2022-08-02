package dtos

import uuid "github.com/satori/go.uuid"

type CreateProductCommandResponse struct {
	ProductID uuid.UUID `json:"productId"`
}
