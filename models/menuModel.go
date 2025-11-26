package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Menu struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	Name      string             `json:"name" validate:"required" example:"Dinner Menu"`
	Category  string             `json:"category" validate:"required" example:"Main Course"`
	StartDate time.Time          `json:"start_date" example:"2024-01-01T00:00:00Z"`
	EndDate   time.Time          `json:"end_date" example:"2024-12-31T23:59:59Z"`
	CreatedAt time.Time          `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time          `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
