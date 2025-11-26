package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Food struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	Name      string             `json:"name" validate:"required,min=2,max=100" example:"Grilled Chicken"`
	Price     float64            `json:"price" validate:"required,gt=0" example:"15.99"`
	FoodImage string             `json:"food_image" validate:"required" example:"https://example.com/images/chicken.jpg"`
	MenuID    string             `json:"menu_id" validate:"required" example:"507f1f77bcf86cd799439011"`
	CreatedAt time.Time          `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time          `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
