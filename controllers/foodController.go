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

func getFoodCollection() *mongo.Collection {
	return database.GetCollection(database.Client, "foods")
}

// @Summary Get All Foods
// @Description Retrieve a complete list of all available food items in the restaurant
// @Tags Food
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.FoodResponse "Array of all food items with details"
// @Failure 401 {object} models.ErrorResponse "Missing or invalid authentication token"
// @Failure 500 {object} models.ErrorResponse "Database error while fetching foods"
// @Router /foods [get]
func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var foods []models.Food
		cursor, err := getFoodCollection().Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching foods"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &foods); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding foods"})
			return
		}

		c.JSON(http.StatusOK, foods)
	}
}

// @Summary Get Food by ID
// @Description Retrieve detailed information about a specific food item using its unique ID
// @Tags Food
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Food MongoDB ObjectID" example("507f1f77bcf86cd799439011")
// @Success 200 {object} models.FoodResponse "Food item details"
// @Failure 400 {object} models.ErrorResponse "Invalid MongoDB ObjectID format"
// @Failure 401 {object} models.ErrorResponse "Missing or invalid authentication token"
// @Failure 404 {object} models.ErrorResponse "Food item not found"
// @Router /foods/{id} [get]
func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		foodID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(foodID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID"})
			return
		}

		var food models.Food
		err = getFoodCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&food)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Food not found"})
			return
		}

		c.JSON(http.StatusOK, food)
	}
}

// @Summary Create Food
// @Description Create a new food item in the restaurant menu (Admin only)
// @Tags Food
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param food body models.FoodCreateRequest true "Food item details (name, price, image, menu_id)"
// @Success 201 {object} models.FoodResponse "Food created successfully with generated ID"
// @Failure 400 {object} models.ErrorResponse "Invalid request body or validation failed"
// @Failure 401 {object} models.ErrorResponse "Missing or invalid authentication token"
// @Failure 500 {object} models.ErrorResponse "Database error while creating food"
// @Router /foods [post]
func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var food models.Food
		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		food.CreatedAt = time.Now()
		food.UpdatedAt = time.Now()
		food.ID = primitive.NewObjectID()

		result, err := getFoodCollection().InsertOne(ctx, food)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create food"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Food created successfully",
			"id":      result.InsertedID,
			"food":    food,
		})
	}
}

// @Summary Update Food
// @Description Update an existing food item's information (Admin only)
// @Tags Food
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Food MongoDB ObjectID" example("507f1f77bcf86cd799439011")
// @Param food body models.FoodCreateRequest true "Updated food details"
// @Success 200 {object} models.SuccessResponse "Food updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid ID format or request body"
// @Failure 401 {object} models.ErrorResponse "Missing or invalid authentication token"
// @Failure 404 {object} models.ErrorResponse "Food item not found"
// @Failure 500 {object} models.ErrorResponse "Database error while updating food"
// @Router /foods/{id} [put]
func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		foodID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(foodID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID"})
			return
		}

		var food models.Food
		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		food.UpdatedAt = time.Now()

		update := bson.M{
			"$set": bson.M{
				"name":       food.Name,
				"price":      food.Price,
				"food_image": food.FoodImage,
				"menu_id":    food.MenuID,
				"updated_at": food.UpdatedAt,
			},
		}

		result, err := getFoodCollection().UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update food"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Food not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Food updated successfully"})
	}
}

// @Summary Delete Food
// @Description Permanently delete a food item from the menu (Admin only)
// @Tags Food
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Food MongoDB ObjectID" example("507f1f77bcf86cd799439011")
// @Success 200 {object} models.SuccessResponse "Food deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid MongoDB ObjectID format"
// @Failure 401 {object} models.ErrorResponse "Missing or invalid authentication token"
// @Failure 404 {object} models.ErrorResponse "Food item not found"
// @Failure 500 {object} models.ErrorResponse "Database error while deleting food"
// @Router /foods/{id} [delete]
func DeleteFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		foodID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(foodID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID"})
			return
		}

		result, err := getFoodCollection().DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete food"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Food not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Food deleted successfully"})
	}
}
