package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	TableNumber    int                `bson:"table_number,omitempty" json:"table_number" validate:"required,min=1" example:"5"`
	Capacity       int                `bson:"capacity,omitempty" json:"capacity" validate:"required,min=1" example:"4"`
	NumberOfGuests int                `bson:"number_of_guests,omitempty" json:"number_of_guests" example:"4"`
	SeatsReserved  int                `bson:"seats_reserved,omitempty" json:"seats_reserved" example:"2"`
	Status         string             `bson:"status,omitempty" json:"status" example:"vacant"`
	IsAvailable    bool               `bson:"is_available,omitempty" json:"is_available" example:"true"`
	CreatedAt      time.Time          `bson:"created_at,omitempty" json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt      time.Time          `bson:"updated_at,omitempty" json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
