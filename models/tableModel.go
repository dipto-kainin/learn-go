package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	TableNumber int                `json:"table_number" validate:"required,min=1" example:"5"`
	Capacity    int                `json:"capacity" validate:"required,min=1" example:"4"`
	IsAvailable bool               `json:"is_available" example:"true"`
	CreatedAt   time.Time          `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time          `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
