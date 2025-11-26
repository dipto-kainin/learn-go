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

func getMenuCollection() *mongo.Collection {
	return database.GetCollection(database.Client, "menus")
}

// @Summary Get All Menus
// @Description Retrieve a list of all menus
// @Tags Menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.MenuResponse "List of menus"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /menus [get]
func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var menus []models.Menu
		cursor, err := getMenuCollection().Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching menus"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &menus); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding menus"})
			return
		}

		c.JSON(http.StatusOK, menus)
	}
}

// @Summary Get Menu by ID
// @Description Retrieve a specific menu by its ID
// @Tags Menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu ID"
// @Success 200 {object} models.MenuResponse "Menu details"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Menu not found"
// @Router /menus/{id} [get]
func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		menuID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(menuID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
			return
		}

		var menu models.Menu
		err = getMenuCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&menu)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		}

		c.JSON(http.StatusOK, menu)
	}
}

// @Summary Create Menu
// @Description Create a new menu
// @Tags Menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param menu body models.MenuCreateRequest true "Menu details"
// @Success 201 {object} models.MenuResponse "Menu created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /menus [post]
func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var menu models.Menu
		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(menu)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		menu.CreatedAt = time.Now()
		menu.UpdatedAt = time.Now()
		menu.ID = primitive.NewObjectID()

		result, err := getMenuCollection().InsertOne(ctx, menu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Menu created successfully",
			"id":      result.InsertedID,
			"menu":    menu,
		})
	}
}

// @Summary Update Menu
// @Description Update an existing menu
// @Tags Menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu ID"
// @Param menu body models.MenuCreateRequest true "Updated menu details"
// @Success 200 {object} models.SuccessResponse "Menu updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 404 {object} models.ErrorResponse "Menu not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /menus/{id} [put]
func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		menuID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(menuID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
			return
		}

		var menu models.Menu
		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		menu.UpdatedAt = time.Now()

		update := bson.M{
			"$set": bson.M{
				"name":       menu.Name,
				"category":   menu.Category,
				"start_date": menu.StartDate,
				"end_date":   menu.EndDate,
				"updated_at": menu.UpdatedAt,
			},
		}

		result, err := getMenuCollection().UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Menu updated successfully"})
	}
}

// @Summary Delete Menu
// @Description Delete a menu by ID
// @Tags Menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu ID"
// @Success 200 {object} models.SuccessResponse "Menu deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Menu not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /menus/{id} [delete]
func DeleteMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		menuID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(menuID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
			return
		}

		result, err := getMenuCollection().DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete menu"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Menu deleted successfully"})
	}
}
