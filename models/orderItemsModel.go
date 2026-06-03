package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	OrderID   string             `bson:"order_id,omitempty" json:"order_id" validate:"required" example:"507f1f77bcf86cd799439012"`
	FoodID    string             `bson:"food_id,omitempty" json:"food_id" validate:"required" example:"507f1f77bcf86cd799439013"`
	Quantity  int                `bson:"quantity,omitempty" json:"quantity" validate:"required,min=1" example:"2"`
	UnitPrice float64            `bson:"unit_price,omitempty" json:"unit_price" validate:"required,gt=0" example:"15.99"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
