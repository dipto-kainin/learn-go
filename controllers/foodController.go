package controllers

import (
	"restroBackend/database"
	"restroBackend/helpers"
	"restroBackend/models"
	"context"
	"fmt"
	"io"
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

		foods := []models.Food{}
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

type FoodMultipartRequest struct {
	Name        string  `form:"name" binding:"required,min=2,max=100"`
	Price       float64 `form:"price" binding:"required,gt=0"`
	MenuID      string  `form:"menu_id" binding:"required"`
	Description string  `form:"description"`
}

// @Summary Create Food
// @Description Create a new food item in the restaurant menu (Admin only)
// @Tags Food
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Food Name"
// @Param price formData number true "Food Price"
// @Param menu_id formData string true "Menu ID"
// @Param description formData string false "Food Description"
// @Param image formData file true "Food Image File"
// @Success 201 {object} models.FoodResponse "Food created successfully with generated ID"
// @Failure 400 {object} models.ErrorResponse "Invalid request body or validation failed"
// @Failure 401 {object} models.ErrorResponse "Missing or invalid authentication token"
// @Failure 500 {object} models.ErrorResponse "Database error while creating food"
// @Router /foods [post]
func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var req FoodMultipartRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		foodID := primitive.NewObjectID()
		var foodURL string
		var imageFileID string

		// Retrieve file from request
		fileHeader, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
			return
		}

		// Read file bytes
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image file"})
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image bytes"})
			return
		}

		// Upload raw bytes to ImageKit
		resp, err := helpers.UploadToImageKit(fileBytes, fmt.Sprintf("food_%s", foodID.Hex()))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image to ImageKit: " + err.Error()})
			return
		}
		foodURL = resp.URL
		imageFileID = resp.FileID

		food := models.Food{
			ID:          foodID,
			Name:        req.Name,
			Price:       req.Price,
			FoodImage:   foodURL,
			ImageFileID: imageFileID,
			MenuID:      req.MenuID,
			Description: req.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

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
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Food MongoDB ObjectID" example("507f1f77bcf86cd799439011")
// @Param name formData string true "Food Name"
// @Param price formData number true "Food Price"
// @Param menu_id formData string true "Menu ID"
// @Param description formData string false "Food Description"
// @Param image formData file false "Food Image File"
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

		var req FoodMultipartRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var existingFood models.Food
		err = getFoodCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&existingFood)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Food not found"})
			return
		}

		foodURL := existingFood.FoodImage
		imageFileID := existingFood.ImageFileID

		// Handle file upload if provided
		fileHeader, err := c.FormFile("image")
		if err == nil && fileHeader != nil {
			// Delete the old image from ImageKit
			if existingFood.ImageFileID != "" {
				_ = helpers.DeleteFromImageKit(existingFood.ImageFileID)
			}

			// Read the file bytes
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read new image file"})
				return
			}
			defer file.Close()

			fileBytes, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read new image bytes"})
				return
			}

			// Upload the new image
			resp, err := helpers.UploadToImageKit(fileBytes, fmt.Sprintf("food_%s", objID.Hex()))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload new image to ImageKit: " + err.Error()})
				return
			}
			foodURL = resp.URL
			imageFileID = resp.FileID
		}

		updatedFood := models.Food{
			ID:          objID,
			Name:        req.Name,
			Price:       req.Price,
			FoodImage:   foodURL,
			ImageFileID: imageFileID,
			MenuID:      req.MenuID,
			Description: req.Description,
			UpdatedAt:   time.Now(),
		}

		validationErr := validate.Struct(updatedFood)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		update := bson.M{
			"$set": bson.M{
				"name":          updatedFood.Name,
				"price":         updatedFood.Price,
				"food_image":    updatedFood.FoodImage,
				"image_file_id": updatedFood.ImageFileID,
				"menu_id":       updatedFood.MenuID,
				"description":   updatedFood.Description,
				"updated_at":    updatedFood.UpdatedAt,
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

		var food models.Food
		err = getFoodCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&food)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Food not found"})
			return
		}

		// Delete from ImageKit if it exists
		if food.ImageFileID != "" {
			_ = helpers.DeleteFromImageKit(food.ImageFileID)
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
