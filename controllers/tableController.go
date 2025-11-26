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

func getTableCollection() *mongo.Collection {
	return database.GetCollection(database.Client, "tables")
}

// @Summary Get All Tables
// @Description Retrieve a list of all restaurant tables
// @Tags Table
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.TableResponse "List of tables"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /tables [get]
func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var tables []models.Table
		cursor, err := getTableCollection().Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching tables"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &tables); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding tables"})
			return
		}

		c.JSON(http.StatusOK, tables)
	}
}

// @Summary Get Table by ID
// @Description Retrieve a specific table by its ID
// @Tags Table
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Table ID"
// @Success 200 {object} models.TableResponse "Table details"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Table not found"
// @Router /tables/{id} [get]
func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		tableID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(tableID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
			return
		}

		var table models.Table
		err = getTableCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}

		c.JSON(http.StatusOK, table)
	}
}

// @Summary Create Table
// @Description Create a new restaurant table
// @Tags Table
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param table body models.TableCreateRequest true "Table details"
// @Success 201 {object} models.TableResponse "Table created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /tables [post]
func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var table models.Table
		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(table)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		table.CreatedAt = time.Now()
		table.UpdatedAt = time.Now()
		table.ID = primitive.NewObjectID()
		table.IsAvailable = true

		result, err := getTableCollection().InsertOne(ctx, table)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Table created successfully",
			"id":      result.InsertedID,
			"table":   table,
		})
	}
}

// @Summary Update Table
// @Description Update an existing table
// @Tags Table
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Table ID"
// @Param table body models.TableCreateRequest true "Updated table details"
// @Success 200 {object} models.SuccessResponse "Table updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 404 {object} models.ErrorResponse "Table not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /tables/{id} [put]
func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		tableID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(tableID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
			return
		}

		var table models.Table
		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		table.UpdatedAt = time.Now()

		update := bson.M{
			"$set": bson.M{
				"table_number": table.TableNumber,
				"capacity":     table.Capacity,
				"is_available": table.IsAvailable,
				"updated_at":   table.UpdatedAt,
			},
		}

		result, err := getTableCollection().UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update table"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Table updated successfully"})
	}
}

// @Summary Delete Table
// @Description Delete a table by ID
// @Tags Table
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Table ID"
// @Success 200 {object} models.SuccessResponse "Table deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Table not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /tables/{id} [delete]
func DeleteTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		tableID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(tableID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
			return
		}

		result, err := getTableCollection().DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete table"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Table deleted successfully"})
	}
}
