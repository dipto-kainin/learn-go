package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	TableID   string             `json:"table_id" validate:"required" example:"507f1f77bcf86cd799439012"`
	OrderDate time.Time          `json:"order_date" example:"2024-01-01T12:00:00Z"`
	Status    string             `json:"status" validate:"required" example:"pending" enums:"pending,preparing,ready,delivered,cancelled"`
	CreatedAt time.Time          `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time          `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
