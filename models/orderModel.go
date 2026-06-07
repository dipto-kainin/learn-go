package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItemNested struct {
	ID        string  `bson:"id" json:"id" validate:"required"`
	Amount    int     `bson:"amount" json:"amount" validate:"required,min=1"`
	UnitPrice float64 `bson:"unit_price" json:"unit_price"`
}

type Order struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	TableID   string             `bson:"table_id,omitempty" json:"table_id" validate:"required" example:"507f1f77bcf86cd799439012"`
	OrderDate time.Time          `bson:"order_date,omitempty" json:"order_date" example:"2024-01-01T12:00:00Z"`
	Status    string             `bson:"status,omitempty" json:"status" validate:"required" example:"pending" enums:"pending,preparing,ready,delivered,cancelled"`
	FoodItems []OrderItemNested  `bson:"food_items" json:"food_items"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
