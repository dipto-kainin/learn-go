package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Booking struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	TableID        string             `bson:"table_id" json:"table_id" validate:"required" example:"507f1f77bcf86cd799439012"`
	UserEmail      string             `bson:"user_email" json:"user_email" validate:"required" example:"customer@example.com"`
	UserName       string             `bson:"user_name" json:"user_name" example:"Jane Doe"`
	UserPhone      string             `bson:"user_phone" json:"user_phone" example:"+919876543210"`
	PartySize      int                `bson:"party_size" json:"party_size" validate:"required,min=1,max=20" example:"4"`
	IsShared       bool               `bson:"is_shared" json:"is_shared" example:"false"`
	ComfortSharing bool               `bson:"comfort_sharing" json:"comfort_sharing" example:"true"`
	Status         string             `bson:"status" json:"status" example:"pending" enums:"pending,checked_in,cancelled"`
	StartTime      time.Time          `bson:"start_time" json:"start_time" validate:"required" example:"2026-06-03T10:00:00Z"`
	EndTime        time.Time          `bson:"end_time" json:"end_time" validate:"required" example:"2026-06-03T12:00:00Z"`
	MinEntryTime   time.Time          `bson:"min_entry_time" json:"min_entry_time" example:"2026-06-03T10:30:00Z"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
