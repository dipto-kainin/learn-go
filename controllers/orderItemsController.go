package controllers

import (
	"basic-backend/database"
	"basic-backend/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getOrderItemCollection() *mongo.Collection {
	return database.GetCollection(database.Client, "orderitems")
}
var validateOrderItem = validator.New()

// @Summary Get Order Items
// @Description Retrieve order items, optionally filtered by order ID
// @Tags OrderItem
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id query string false "Filter by Order ID"
// @Success 200 {array} models.OrderItemResponse "List of order items"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orderitems [get]
func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Query("order_id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{}
		if orderID != "" {
			filter["order_id"] = orderID
		}

		var orderItems []models.OrderItem
		cursor, err := getOrderItemCollection().Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching order items"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &orderItems); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding order items"})
			return
		}

		c.JSON(http.StatusOK, orderItems)
	}
}

// @Summary Get Order Item by ID
// @Description Retrieve a specific order item by its ID
// @Tags OrderItem
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order Item ID"
// @Success 200 {object} models.OrderItemResponse "Order item details"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Order item not found"
// @Router /orderitems/{id} [get]
func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderItemID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(orderItemID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order item ID"})
			return
		}

		var orderItem models.OrderItem
		err = getOrderItemCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&orderItem)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order item not found"})
			return
		}

		c.JSON(http.StatusOK, orderItem)
	}
}

// @Summary Create Order Item
// @Description Create a new order item
// @Tags OrderItem
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orderitem body models.OrderItemCreateRequest true "Order item details"
// @Success 201 {object} models.OrderItemResponse "Order item created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orderitems [post]
func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var orderItem models.OrderItem
		if err := c.BindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validateOrderItem.Struct(orderItem)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		orderItem.CreatedAt = time.Now()
		orderItem.UpdatedAt = time.Now()
		orderItem.ID = primitive.NewObjectID()

		result, err := getOrderItemCollection().InsertOne(ctx, orderItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order item"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":    "Order item created successfully",
			"id":         result.InsertedID,
			"order_item": orderItem,
		})
	}
}

// @Summary Update Order Item
// @Description Update an existing order item
// @Tags OrderItem
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order Item ID"
// @Param orderitem body models.OrderItemCreateRequest true "Updated order item details"
// @Success 200 {object} models.SuccessResponse "Order item updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 404 {object} models.ErrorResponse "Order item not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orderitems/{id} [put]
func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderItemID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(orderItemID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order item ID"})
			return
		}

		var orderItem models.OrderItem
		if err := c.BindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orderItem.UpdatedAt = time.Now()

		update := bson.M{
			"$set": bson.M{
				"order_id":   orderItem.OrderID,
				"food_id":    orderItem.FoodID,
				"quantity":   orderItem.Quantity,
				"unit_price": orderItem.UnitPrice,
				"updated_at": orderItem.UpdatedAt,
			},
		}

		result, err := getOrderItemCollection().UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order item"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order item not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order item updated successfully"})
	}
}

// @Summary Delete Order Item
// @Description Delete an order item by ID
// @Tags OrderItem
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order Item ID"
// @Success 200 {object} models.SuccessResponse "Order item deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Order item not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orderitems/{id} [delete]
func DeleteOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderItemID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(orderItemID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order item ID"})
			return
		}

		result, err := getOrderItemCollection().DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order item"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order item not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order item deleted successfully"})
	}
}
