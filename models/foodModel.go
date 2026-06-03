package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Food struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	Name        string             `json:"name" validate:"required,min=2,max=100" example:"Grilled Chicken"`
	Price       float64            `json:"price" validate:"required,gt=0" example:"150.00"`
	FoodImage   string             `bson:"food_image,omitempty" json:"food_image" validate:"required" example:"https://example.com/images/chicken.jpg"`
	ImageFileID string             `bson:"image_file_id,omitempty" json:"image_file_id" example:"file_id_here"`
	MenuID      string             `json:"menu_id" validate:"required" example:"507f1f77bcf86cd799439011"`
	Description string             `bson:"description,omitempty" json:"description" example:"Delicious dish"`
	ImageBase64 string             `bson:"-" json:"image_base64,omitempty"` // Base64 data for photo uploads
	CreatedAt   time.Time          `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time          `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
