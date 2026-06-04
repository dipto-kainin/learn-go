package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	FirstName    string             `bson:"first_name,omitempty" json:"first_name" validate:"required,min=2,max=100" example:"John"`
	LastName     string             `bson:"last_name,omitempty" json:"last_name" validate:"required,min=2,max=100" example:"Doe"`
	Email        string             `bson:"email,omitempty" json:"email" validate:"email,required" example:"john.doe@example.com"`
	Password     string             `bson:"password,omitempty" json:"password" validate:"required,min=6" example:"password123"`
	Phone        string             `bson:"phone,omitempty" json:"phone" validate:"required" example:"+1234567890"`
	Token        string             `bson:"token,omitempty" json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string             `bson:"refresh_token,omitempty" json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	CreatedAt    time.Time          `bson:"created_at,omitempty" json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty" json:"updated_at" example:"2024-01-01T00:00:00Z"`
	UserType     string             `bson:"user_type,omitempty" json:"user_type" validate:"required,eq=ADMIN|eq=STAFF|eq=USER" example:"USER" enums:"ADMIN,STAFF,USER"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}
