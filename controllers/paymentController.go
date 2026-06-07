package controllers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"restroBackend/database"
	"restroBackend/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// @Summary Create Razorpay Order
// @Description Initiate a Razorpay payment order for checkout
// @Tags Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string]string true "Local Order ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /payments/create-order [post]
func CreateRazorpayOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var req struct {
			OrderID string `json:"order_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Calculate total amount for the order from embedded FoodItems
		objID, err := primitive.ObjectIDFromHex(req.OrderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		var order models.Order
		orderCol := database.GetCollection(database.Client, "orders")
		err = orderCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		var total float64
		for _, item := range order.FoodItems {
			total += float64(item.Amount) * item.UnitPrice
		}

		if total <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot checkout an empty order"})
			return
		}

		// Call Razorpay Order API
		client := &http.Client{Timeout: 10 * time.Second}
		razorpayKey := os.Getenv("RAZOR_PAY_KEY")
		razorpaySecret := os.Getenv("RAZOR_PAY_SECRET")

		payload := map[string]interface{}{
			"amount":   int64(total * 100), // amount in paise
			"currency": "INR",
			"receipt":  req.OrderID,
		}
		jsonPayload, _ := json.Marshal(payload)

		rReq, err := http.NewRequest("POST", "https://api.razorpay.com/v1/orders", bytes.NewBuffer(jsonPayload))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order request"})
			return
		}
		rReq.Header.Set("Content-Type", "application/json")
		rReq.SetBasicAuth(razorpayKey, razorpaySecret)

		resp, err := client.Do(rReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reach Razorpay API"})
			return
		}
		defer resp.Body.Close()

		var razorpayResp struct {
			ID       string `json:"id"`
			Amount   int64  `json:"amount"`
			Currency string `json:"currency"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&razorpayResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode Razorpay response"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":       razorpayResp.ID,
			"amount":   razorpayResp.Amount,
			"currency": razorpayResp.Currency,
			"key":      razorpayKey,
		})
	}
}

// @Summary Verify Razorpay Payment Signature
// @Description Validate payment success, generate invoice, and free table
// @Tags Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body map[string]string true "Razorpay credentials"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /payments/verify [post]
func VerifyRazorpayPayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var req struct {
			OrderID           string `json:"order_id" binding:"required"`
			RazorpayPaymentID string `json:"razorpay_payment_id" binding:"required"`
			RazorpayOrderID   string `json:"razorpay_order_id" binding:"required"`
			RazorpaySignature string `json:"razorpay_signature" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Perform cryptographic signature check
		razorpaySecret := os.Getenv("RAZOR_PAY_SECRET")
		message := req.RazorpayOrderID + "|" + req.RazorpayPaymentID
		h := hmac.New(sha256.New, []byte(razorpaySecret))
		h.Write([]byte(message))
		generatedSignature := hex.EncodeToString(h.Sum(nil))

		if generatedSignature != req.RazorpaySignature {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payment signature verification failed"})
			return
		}

		// Signature matches - perform invoice creation & status changes
		objID, err := primitive.ObjectIDFromHex(req.OrderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		var order models.Order
		orderCol := database.GetCollection(database.Client, "orders")
		err = orderCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		// Calculate total order amount from embedded FoodItems
		var total float64
		for _, item := range order.FoodItems {
			total += float64(item.Amount) * item.UnitPrice
		}

		// Save invoice
		invoiceCol := database.GetCollection(database.Client, "invoices")
		var existingInvoice models.Invoice
		invoiceErr := invoiceCol.FindOne(ctx, bson.M{"order_id": req.OrderID}).Decode(&existingInvoice)

		if invoiceErr == mongo.ErrNoDocuments {
			invoice := models.Invoice{
				ID:            primitive.NewObjectID(),
				OrderID:       req.OrderID,
				PaymentMethod: "upi",
				TotalAmount:   total,
				PaymentStatus: "paid",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			_, err = invoiceCol.InsertOne(ctx, invoice)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save invoice"})
				return
			}
		}

		// Update order status to completed
		_, err = orderCol.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
			"$set": bson.M{
				"status":     "completed",
				"updated_at": time.Now(),
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
			return
		}

		// Update table status to vacant
		tableCol := database.GetCollection(database.Client, "tables")
		tableObjID, err := primitive.ObjectIDFromHex(order.TableID)
		if err == nil {
			_, _ = tableCol.UpdateOne(ctx, bson.M{"_id": tableObjID}, bson.M{
				"$set": bson.M{
					"status":       "vacant",
					"is_available": true,
					"updated_at":   time.Now(),
				},
			})

			// Complete bookings associated with the table
			bookingCol := database.GetCollection(database.Client, "bookings")
			now := time.Now()
			bCursor, bErr := bookingCol.Find(ctx, bson.M{
				"table_id": order.TableID,
				"status":   bson.M{"$in": []string{"pending", "checked_in"}},
			})
			if bErr == nil {
				var tableBookings []models.Booking
				if err := bCursor.All(ctx, &tableBookings); err == nil {
					for _, b := range tableBookings {
						if b.Status == "checked_in" || b.StartTime.Before(now) || b.StartTime.IsZero() {
							newStatus := "completed"
							if b.Status == "pending" {
								newStatus = "cancelled"
							}
							_, _ = bookingCol.UpdateOne(ctx, bson.M{"_id": b.ID}, bson.M{
								"$set": bson.M{
									"status":     newStatus,
									"updated_at": now,
								},
							})
						}
					}
				}
				bCursor.Close(ctx)
			}

			// Ensure full floor map consistency
			updateTableStatusForCurrentTime(ctx, order.TableID)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payment verified and order finalized"})
	}
}

// @Summary Request Cash Payment
// @Description Initiate a cash checkout request by the customer
// @Tags Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string]string true "Local Order ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /payments/request-cash [post]
func RequestCashPayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var req struct {
			OrderID string `json:"order_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		objID, err := primitive.ObjectIDFromHex(req.OrderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		var order models.Order
		orderCol := database.GetCollection(database.Client, "orders")
		err = orderCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		var total float64
		for _, item := range order.FoodItems {
			total += float64(item.Amount) * item.UnitPrice
		}

		if total <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot checkout an empty order"})
			return
		}

		// Save invoice with payment_status: pending
		invoiceCol := database.GetCollection(database.Client, "invoices")
		var existingInvoice models.Invoice
		invoiceErr := invoiceCol.FindOne(ctx, bson.M{"order_id": req.OrderID}).Decode(&existingInvoice)

		if invoiceErr == mongo.ErrNoDocuments {
			invoice := models.Invoice{
				ID:            primitive.NewObjectID(),
				OrderID:       req.OrderID,
				PaymentMethod: "cash",
				TotalAmount:   total,
				PaymentStatus: "pending",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			_, err = invoiceCol.InsertOne(ctx, invoice)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to request cash payment"})
				return
			}
		} else {
			// Update existing invoice to cash/pending
			_, err = invoiceCol.UpdateOne(ctx, bson.M{"order_id": req.OrderID}, bson.M{
				"$set": bson.M{
					"payment_method": "cash",
					"payment_status": "pending",
					"updated_at":     time.Now(),
				},
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cash payment request"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cash payment requested. Please wait for staff to collect cash."})
	}
}

// @Summary Confirm Cash Payment
// @Description Staff confirms physical cash collection and vacates table
// @Tags Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string]string true "Local Order ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /payments/confirm-cash [post]
func ConfirmCashPayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var req struct {
			OrderID string `json:"order_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		objID, err := primitive.ObjectIDFromHex(req.OrderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		var order models.Order
		orderCol := database.GetCollection(database.Client, "orders")
		err = orderCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		var total float64
		for _, item := range order.FoodItems {
			total += float64(item.Amount) * item.UnitPrice
		}

		// Save/Update invoice to status paid
		invoiceCol := database.GetCollection(database.Client, "invoices")
		var existingInvoice models.Invoice
		invoiceErr := invoiceCol.FindOne(ctx, bson.M{"order_id": req.OrderID}).Decode(&existingInvoice)

		if invoiceErr == mongo.ErrNoDocuments {
			invoice := models.Invoice{
				ID:            primitive.NewObjectID(),
				OrderID:       req.OrderID,
				PaymentMethod: "cash",
				TotalAmount:   total,
				PaymentStatus: "paid",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			_, err = invoiceCol.InsertOne(ctx, invoice)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save invoice"})
				return
			}
		} else {
			_, err = invoiceCol.UpdateOne(ctx, bson.M{"order_id": req.OrderID}, bson.M{
				"$set": bson.M{
					"payment_status": "paid",
					"payment_method": "cash",
					"updated_at":     time.Now(),
				},
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize invoice"})
				return
			}
		}

		// Update order status to completed
		_, err = orderCol.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
			"$set": bson.M{
				"status":     "completed",
				"updated_at": time.Now(),
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
			return
		}

		// Update table status to vacant
		tableCol := database.GetCollection(database.Client, "tables")
		tableObjID, err := primitive.ObjectIDFromHex(order.TableID)
		if err == nil {
			_, _ = tableCol.UpdateOne(ctx, bson.M{"_id": tableObjID}, bson.M{
				"$set": bson.M{
					"status":       "vacant",
					"is_available": true,
					"updated_at":   time.Now(),
				},
			})

			// Complete bookings associated with the table
			bookingCol := database.GetCollection(database.Client, "bookings")
			now := time.Now()
			bCursor, bErr := bookingCol.Find(ctx, bson.M{
				"table_id": order.TableID,
				"status":   bson.M{"$in": []string{"pending", "checked_in"}},
			})
			if bErr == nil {
				var tableBookings []models.Booking
				if err := bCursor.All(ctx, &tableBookings); err == nil {
					for _, b := range tableBookings {
						if b.Status == "checked_in" || b.StartTime.Before(now) || b.StartTime.IsZero() {
							newStatus := "completed"
							if b.Status == "pending" {
								newStatus = "cancelled"
							}
							_, _ = bookingCol.UpdateOne(ctx, bson.M{"_id": b.ID}, bson.M{
								"$set": bson.M{
									"status":     newStatus,
									"updated_at": now,
								},
							})
						}
					}
				}
				bCursor.Close(ctx)
			}

			// Ensure full floor map consistency
			updateTableStatusForCurrentTime(ctx, order.TableID)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cash payment confirmed and table finalized"})
	}
}
