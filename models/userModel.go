package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	FirstName    string             `json:"first_name" validate:"required,min=2,max=100" example:"John"`
	LastName     string             `json:"last_name" validate:"required,min=2,max=100" example:"Doe"`
	Email        string             `json:"email" validate:"email,required" example:"john.doe@example.com"`
	Password     string             `json:"password" validate:"required,min=6" example:"password123"`
	Phone        string             `json:"phone" validate:"required" example:"+1234567890"`
	Token        string             `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string             `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	CreatedAt    time.Time          `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt    time.Time          `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	UserType     string             `json:"user_type" validate:"required,eq=ADMIN|eq=USER" example:"USER" enums:"USER,ADMIN"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"email,required" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}
