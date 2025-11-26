package controllers

import (
	"basic-backend/database"
	"basic-backend/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getInvoiceCollection() *mongo.Collection {
	return database.GetCollection(database.Client, "invoices")
}

// @Summary Get All Invoices
// @Description Retrieve a list of all invoices
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.InvoiceResponse "List of invoices"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /invoices [get]
func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var invoices []models.Invoice
		cursor, err := getInvoiceCollection().Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching invoices"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &invoices); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding invoices"})
			return
		}

		c.JSON(http.StatusOK, invoices)
	}
}

// @Summary Get Invoice by ID
// @Description Retrieve a specific invoice by its ID
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {object} models.InvoiceResponse "Invoice details"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Invoice not found"
// @Router /invoices/{id} [get]
func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		invoiceID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(invoiceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
			return
		}

		var invoice models.Invoice
		err = getInvoiceCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&invoice)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
			return
		}

		c.JSON(http.StatusOK, invoice)
	}
}

// @Summary Create Invoice
// @Description Create a new invoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param invoice body models.InvoiceCreateRequest true "Invoice details"
// @Success 201 {object} models.InvoiceResponse "Invoice created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /invoices [post]
func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var invoice models.Invoice
		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		invoice.CreatedAt = time.Now()
		invoice.UpdatedAt = time.Now()
		invoice.ID = primitive.NewObjectID()

		result, err := getInvoiceCollection().InsertOne(ctx, invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invoice"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Invoice created successfully",
			"id":      result.InsertedID,
			"invoice": invoice,
		})
	}
}

// @Summary Update Invoice
// @Description Update an existing invoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Param invoice body models.InvoiceCreateRequest true "Updated invoice details"
// @Success 200 {object} models.SuccessResponse "Invoice updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 404 {object} models.ErrorResponse "Invoice not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /invoices/{id} [put]
func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		invoiceID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(invoiceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
			return
		}

		var invoice models.Invoice
		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		invoice.UpdatedAt = time.Now()

		update := bson.M{
			"$set": bson.M{
				"order_id":       invoice.OrderID,
				"payment_method": invoice.PaymentMethod,
				"total_amount":   invoice.TotalAmount,
				"payment_status": invoice.PaymentStatus,
				"updated_at":     invoice.UpdatedAt,
			},
		}

		result, err := getInvoiceCollection().UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Invoice updated successfully"})
	}
}

