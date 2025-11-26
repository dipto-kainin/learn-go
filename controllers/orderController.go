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

func getOrderCollection() *mongo.Collection {
	return database.GetCollection(database.Client, "orders")
}
var validateOrder = validator.New()

// @Summary Get All Orders
// @Description Retrieve a list of all orders
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.OrderResponse "List of orders"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orders [get]
func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var orders []models.Order
		cursor, err := getOrderCollection().Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching orders"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &orders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding orders"})
			return
		}

		c.JSON(http.StatusOK, orders)
	}
}

// @Summary Get Order by ID
// @Description Retrieve a specific order by its ID
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} models.OrderResponse "Order details"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Order not found"
// @Router /orders/{id} [get]
func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(orderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		var order models.Order
		err = getOrderCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

// @Summary Create Order
// @Description Create a new order
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body models.OrderCreateRequest true "Order details"
// @Success 201 {object} models.OrderResponse "Order created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orders [post]
func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var order models.Order
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validateOrder.Struct(order)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		order.CreatedAt = time.Now()
		order.UpdatedAt = time.Now()
		order.OrderDate = time.Now()
		order.ID = primitive.NewObjectID()

		result, err := getOrderCollection().InsertOne(ctx, order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Order created successfully",
			"id":      result.InsertedID,
			"order":   order,
		})
	}
}

// @Summary Update Order
// @Description Update an existing order
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param order body models.OrderCreateRequest true "Updated order details"
// @Success 200 {object} models.SuccessResponse "Order updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 404 {object} models.ErrorResponse "Order not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orders/{id} [put]
func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(orderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		var order models.Order
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.UpdatedAt = time.Now()

		update := bson.M{
			"$set": bson.M{
				"table_id":   order.TableID,
				"status":     order.Status,
				"updated_at": order.UpdatedAt,
			},
		}

		result, err := getOrderCollection().UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully"})
	}
}

// @Summary Delete Order
// @Description Delete an order by ID
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} models.SuccessResponse "Order deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Order not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orders/{id} [delete]
func DeleteOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(orderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		result, err := getOrderCollection().DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
	}
}

