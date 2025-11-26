package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Invoice struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	OrderID       string             `json:"order_id" validate:"required" example:"507f1f77bcf86cd799439012"`
	PaymentMethod string             `json:"payment_method" validate:"required" example:"credit_card" enums:"cash,credit_card,debit_card,mobile_payment"`
	TotalAmount   float64            `json:"total_amount" validate:"required,gt=0" example:"45.99"`
	PaymentStatus string             `json:"payment_status" validate:"required" example:"paid" enums:"pending,paid,failed,refunded"`
	CreatedAt     time.Time          `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt     time.Time          `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
