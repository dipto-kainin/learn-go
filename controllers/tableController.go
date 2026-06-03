package controllers

import (
	"context"
	"net/http"
	"restroBackend/database"
	"restroBackend/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getTableCollection() *mongo.Collection {
	return database.GetCollection(database.Client, "tables")
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006-01-02T15:04:05", s)
	return t, err
}

func enrichTableForInterval(ctx context.Context, table *models.Table, start time.Time, end time.Time) {
	bookingCol := database.GetCollection(database.Client, "bookings")

	endBuf := end.Add(30 * time.Minute)
	filter := bson.M{
		"table_id":   table.ID.Hex(),
		"status":     bson.M{"$in": []string{"pending", "checked_in"}},
		"start_time": bson.M{"$lt": endBuf},
		"end_time":   bson.M{"$gt": start.Add(-30 * time.Minute)},
	}

	cursor, err := bookingCol.Find(ctx, filter)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	if err = cursor.All(ctx, &bookings); err != nil {
		return
	}

	if len(bookings) == 0 {
		if table.Status != "occupied" {
			table.Status = "vacant"
			table.IsAvailable = true
		}
		table.SeatsReserved = 0
		return
	}

	hasExclusive := false
	hasCheckedIn := false
	for _, b := range bookings {
		if !b.IsShared {
			hasExclusive = true
		}
		if b.Status == "checked_in" {
			hasCheckedIn = true
		}
	}

	if hasExclusive {
		table.SeatsReserved = table.Capacity
		table.IsAvailable = false
		if table.Status != "occupied" {
			if hasCheckedIn {
				table.Status = "occupied"
			} else {
				table.Status = "reserved"
			}
		}
		return
	}

	timePoints := map[time.Time]bool{
		start: true,
	}
	for _, b := range bookings {
		bStart := b.StartTime
		bEndBuf := b.EndTime.Add(30 * time.Minute)
		if bStart.After(start) && bStart.Before(endBuf) {
			timePoints[bStart] = true
		}
		if bEndBuf.After(start) && bEndBuf.Before(endBuf) {
			timePoints[bEndBuf] = true
		}
	}

	peakSeats := 0
	for tp := range timePoints {
		occupiedAtTP := 0
		for _, b := range bookings {
			bStart := b.StartTime
			bEndBuf := b.EndTime.Add(30 * time.Minute)
			if (bStart.Before(tp) || bStart.Equal(tp)) && tp.Before(bEndBuf) {
				occupiedAtTP += b.PartySize
			}
		}
		if occupiedAtTP > peakSeats {
			peakSeats = occupiedAtTP
		}
	}

	table.SeatsReserved = peakSeats
	table.IsAvailable = table.SeatsReserved < table.Capacity

	if table.Status != "occupied" {
		if hasCheckedIn {
			table.Status = "occupied"
		} else {
			table.Status = "reserved"
		}
	}
}

func updateTableStatusForCurrentTime(ctx context.Context, tableID string) {
	tableObjID, err := primitive.ObjectIDFromHex(tableID)
	if err != nil {
		return
	}

	var table models.Table
	tableCol := getTableCollection()
	err = tableCol.FindOne(ctx, bson.M{"_id": tableObjID}).Decode(&table)
	if err != nil {
		return
	}

	now := time.Now()
	enrichTableForInterval(ctx, &table, now, now.Add(2*time.Hour))

	update := bson.M{
		"$set": bson.M{
			"status":         table.Status,
			"is_available":   table.IsAvailable,
			"seats_reserved": table.SeatsReserved,
			"updated_at":     now,
		},
	}
	_, _ = tableCol.UpdateOne(ctx, bson.M{"_id": tableObjID}, update)
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

		tables := []models.Table{}
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

		startTimeStr := c.Query("start_time")
		endTimeStr := c.Query("end_time")

		var start, end time.Time
		var parseErr error
		if startTimeStr != "" && endTimeStr != "" {
			start, parseErr = parseTime(startTimeStr)
			if parseErr == nil {
				end, parseErr = parseTime(endTimeStr)
			}
		}

		if startTimeStr == "" || endTimeStr == "" || parseErr != nil {
			start = time.Now()
			end = time.Now().Add(2 * time.Hour)
		}

		for i := range tables {
			enrichTableForInterval(ctx, &tables[i], start, end)
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

		startTimeStr := c.Query("start_time")
		endTimeStr := c.Query("end_time")

		var start, end time.Time
		var parseErr error
		if startTimeStr != "" && endTimeStr != "" {
			start, parseErr = parseTime(startTimeStr)
			if parseErr == nil {
				end, parseErr = parseTime(endTimeStr)
			}
		}

		if startTimeStr == "" || endTimeStr == "" || parseErr != nil {
			start = time.Now()
			end = time.Now().Add(2 * time.Hour)
		}

		enrichTableForInterval(ctx, &table, start, end)

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

		if table.Capacity == 0 && table.NumberOfGuests > 0 {
			table.Capacity = table.NumberOfGuests
		}
		if table.NumberOfGuests == 0 && table.Capacity > 0 {
			table.NumberOfGuests = table.Capacity
		}
		if table.Status == "" {
			table.Status = "vacant"
		}
		table.IsAvailable = (table.Status == "vacant")

		validationErr := validate.Struct(table)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		// Check if table number already exists
		count, err := getTableCollection().CountDocuments(ctx, bson.M{"table_number": table.TableNumber})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking table number uniqueness"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Table number already exists"})
			return
		}

		table.CreatedAt = time.Now()
		table.UpdatedAt = time.Now()
		table.ID = primitive.NewObjectID()

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

		if table.Capacity == 0 && table.NumberOfGuests > 0 {
			table.Capacity = table.NumberOfGuests
		}
		if table.NumberOfGuests == 0 && table.Capacity > 0 {
			table.NumberOfGuests = table.Capacity
		}
		if table.Status != "" {
			table.IsAvailable = (table.Status == "vacant")
		} else {
			table.Status = "occupied"
			if table.IsAvailable {
				table.Status = "vacant"
			}
		}

		// Check if table number already exists on a different table
		count, err := getTableCollection().CountDocuments(ctx, bson.M{
			"table_number": table.TableNumber,
			"_id":          bson.M{"$ne": objID},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking table number uniqueness"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Table number already exists"})
			return
		}

		table.UpdatedAt = time.Now()

		update := bson.M{
			"$set": bson.M{
				"table_number":     table.TableNumber,
				"capacity":         table.Capacity,
				"number_of_guests": table.NumberOfGuests,
				"status":           table.Status,
				"is_available":     table.IsAvailable,
				"updated_at":       table.UpdatedAt,
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

		if table.Status == "vacant" {
			bookingCol := database.GetCollection(database.Client, "bookings")
			now := time.Now()
			cursor, err := bookingCol.Find(ctx, bson.M{
				"table_id": tableID,
				"status":   bson.M{"$in": []string{"pending", "checked_in"}},
			})
			if err == nil {
				var tableBookings []models.Booking
				if err := cursor.All(ctx, &tableBookings); err == nil {
					for _, b := range tableBookings {
						if b.StartTime.Before(now) || b.StartTime.IsZero() {
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
				cursor.Close(ctx)
			}
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
