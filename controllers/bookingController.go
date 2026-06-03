package controllers

import (
	"context"
	"fmt"
	"net/http"
	"restroBackend/database"
	"restroBackend/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getBookingCollection() *mongo.Collection {
	return database.GetCollection(database.Client, "bookings")
}

// @Summary Create Booking
// @Description Customer creates a table booking (soft lock / reservation)
// @Tags Booking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param booking body models.BookingCreateRequest true "Booking details"
// @Success 201 {object} models.BookingResponse "Booking created successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 409 {object} models.ErrorResponse "Table not available"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /bookings [post]
func CreateBooking() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var booking models.Booking
		if err := c.BindJSON(&booking); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if booking.PartySize < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Party size must be at least 1"})
			return
		}

		if booking.StartTime.IsZero() || booking.EndTime.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start time and end time are required"})
			return
		}

		if booking.EndTime.Before(booking.StartTime) || booking.EndTime.Equal(booking.StartTime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "End time must be after start time"})
			return
		}

		if booking.StartTime.Before(time.Now().Add(-1 * time.Minute)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Booking start time must be in the future"})
			return
		}

		// Get the user info from auth context
		userType := c.GetString("user_type")
		if userType == "USER" || booking.UserEmail == "" {
			booking.UserEmail = c.GetString("email")
			booking.UserName = c.GetString("first_name") + " " + c.GetString("last_name")
		}

		// Validate table exists
		tableObjID, err := primitive.ObjectIDFromHex(booking.TableID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
			return
		}

		var table models.Table
		err = getTableCollection().FindOne(ctx, bson.M{"_id": tableObjID}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}

		if table.Status == "occupied" && !booking.IsShared {
			now := time.Now()
			if booking.StartTime.Before(now.Add(2 * time.Hour)) && booking.EndTime.After(now) {
				c.JSON(http.StatusConflict, gin.H{"error": "Table is currently occupied. New reservations must be shared seating."})
				return
			}
		}

		// Get requested interval with 30-min buffer
		reqStart := booking.StartTime
		reqEndBuf := booking.EndTime.Add(30 * time.Minute)

		// Calculate currently reserved seats from active bookings on this table during the requested interval
		var activeBookings []models.Booking
		cursor, err := getBookingCollection().Find(ctx, bson.M{
			"table_id": booking.TableID,
			"status":   bson.M{"$in": []string{"pending", "checked_in"}},
		})
		if err == nil {
			cursor.All(ctx, &activeBookings)
			cursor.Close(ctx)
		}

		// Find active bookings that overlap with [reqStart, reqEndBuf]
		var overlappingBookings []models.Booking
		for _, ab := range activeBookings {
			abStart := ab.StartTime
			abEndBuf := ab.EndTime.Add(30 * time.Minute)
			if reqStart.Before(abEndBuf) && abStart.Before(reqEndBuf) {
				overlappingBookings = append(overlappingBookings, ab)
			}
		}

		// Exclusivity checks
		for _, ob := range overlappingBookings {
			if !ob.IsShared {
				c.JSON(http.StatusConflict, gin.H{"error": "Table is exclusively booked during this time slot."})
				return
			}
		}

		if !booking.IsShared && len(overlappingBookings) > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Table already has bookings during this slot. Cannot book exclusively."})
			return
		}

		// Check capacity limits at all critical time boundary events in the interval
		if booking.IsShared {
			timePoints := map[time.Time]bool{
				reqStart: true,
			}
			for _, ob := range overlappingBookings {
				obStart := ob.StartTime
				obEndBuf := ob.EndTime.Add(30 * time.Minute)
				if obStart.After(reqStart) && obStart.Before(reqEndBuf) {
					timePoints[obStart] = true
				}
				if obEndBuf.After(reqStart) && obEndBuf.Before(reqEndBuf) {
					timePoints[obEndBuf] = true
				}
			}

			for tp := range timePoints {
				occupiedAtTP := 0
				for _, ob := range overlappingBookings {
					obStart := ob.StartTime
					obEndBuf := ob.EndTime.Add(30 * time.Minute)
					if (obStart.Before(tp) || obStart.Equal(tp)) && tp.Before(obEndBuf) {
						occupiedAtTP += ob.PartySize
					}
				}
				if occupiedAtTP+booking.PartySize > table.Capacity {
					c.JSON(http.StatusConflict, gin.H{
						"error": "Not enough seats available during the selected time period.",
					})
					return
				}
			}
		} else {
			if booking.PartySize > table.Capacity {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Party size exceeds table capacity"})
				return
			}
		}

		// Create the booking
		booking.Status = "pending"
		booking.CreatedAt = time.Now()
		booking.UpdatedAt = time.Now()
		booking.ID = primitive.NewObjectID()

		// Calculate default minimum entry time (cutoff time): floor(duration) * 30 min (min 30 min)
		durationHours := booking.EndTime.Sub(booking.StartTime).Hours()
		limitMinutes := int(durationHours) * 30
		if limitMinutes < 30 {
			limitMinutes = 30
		}
		booking.MinEntryTime = booking.StartTime.Add(time.Duration(limitMinutes) * time.Minute)

		_, insertErr := getBookingCollection().InsertOne(ctx, booking)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
			return
		}

		// Update table status in DB for the current instant
		updateTableStatusForCurrentTime(ctx, booking.TableID)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Booking created successfully",
			"id":      booking.ID,
			"booking": booking,
		})
	}
}

// @Summary Get All Bookings
// @Description Retrieve bookings. Customers see their own bookings, staff/admin see all.
// @Tags Booking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Booking "List of bookings"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /bookings [get]
func GetBookings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{}
		userType := c.GetString("user_type")
		if userType == "USER" {
			// Customers only see their own bookings
			filter["user_email"] = c.GetString("email")
		}

		bookings := []models.Booking{}
		cursor, err := getBookingCollection().Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching bookings"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &bookings); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding bookings"})
			return
		}

		c.JSON(http.StatusOK, bookings)
	}
}

// @Summary Get Booking by ID
// @Description Retrieve a specific booking by its ID
// @Tags Booking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Success 200 {object} models.Booking "Booking details"
// @Failure 400 {object} models.ErrorResponse "Invalid ID"
// @Failure 404 {object} models.ErrorResponse "Booking not found"
// @Router /bookings/{id} [get]
func GetBooking() gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(bookingID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
			return
		}

		var booking models.Booking
		err = getBookingCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&booking)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		// Customers can only view their own bookings
		userType := c.GetString("user_type")
		if userType == "USER" && booking.UserEmail != c.GetString("email") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		c.JSON(http.StatusOK, booking)
	}
}

// @Summary Check In Booking
// @Description Staff/admin verifies and checks in a customer booking (hard lock)
// @Tags Booking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Success 200 {object} models.SuccessResponse "Booking checked in"
// @Failure 400 {object} models.ErrorResponse "Invalid ID or booking not pending"
// @Failure 404 {object} models.ErrorResponse "Booking not found"
// @Router /bookings/{id}/check-in [put]
func CheckInBooking() gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(bookingID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
			return
		}

		var booking models.Booking
		err = getBookingCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&booking)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if booking.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending bookings can be checked in"})
			return
		}

		// Load table details to verify capacity
		tableObjID, err := primitive.ObjectIDFromHex(booking.TableID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID on booking"})
			return
		}
		var table models.Table
		err = getTableCollection().FindOne(ctx, bson.M{"_id": tableObjID}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}

		// Calculate check-in dining interval [now, now + duration + 30 min]
		now := time.Now()
		duration := booking.EndTime.Sub(booking.StartTime)
		checkInEndBuf := now.Add(duration).Add(30 * time.Minute)

		// Query OTHER active bookings on this table during this period
		var otherBookings []models.Booking
		cursor, err := getBookingCollection().Find(ctx, bson.M{
			"table_id": booking.TableID,
			"status":   bson.M{"$in": []string{"pending", "checked_in"}},
			"_id":      bson.M{"$ne": objID},
		})
		if err == nil {
			cursor.All(ctx, &otherBookings)
			cursor.Close(ctx)
		}

		var overlapping []models.Booking
		for _, ob := range otherBookings {
			obStart := ob.StartTime
			obEndBuf := ob.EndTime.Add(30 * time.Minute)
			if now.Before(obEndBuf) && obStart.Before(checkInEndBuf) {
				overlapping = append(overlapping, ob)
			}
		}

		// Exclusivity and capacity checks
		if !booking.IsShared && len(overlapping) > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot check in: table has other bookings scheduled during this dining window."})
			return
		}

		for _, ob := range overlapping {
			if !ob.IsShared {
				c.JSON(http.StatusConflict, gin.H{"error": "Cannot check in: table has an exclusive booking scheduled during this dining window."})
				return
			}
		}

		if booking.IsShared {
			timePoints := map[time.Time]bool{
				now: true,
			}
			for _, ob := range overlapping {
				obStart := ob.StartTime
				obEndBuf := ob.EndTime.Add(30 * time.Minute)
				if obStart.After(now) && obStart.Before(checkInEndBuf) {
					timePoints[obStart] = true
				}
				if obEndBuf.After(now) && obEndBuf.Before(checkInEndBuf) {
					timePoints[obEndBuf] = true
				}
			}

			for tp := range timePoints {
				occupiedAtTP := 0
				for _, ob := range overlapping {
					obStart := ob.StartTime
					obEndBuf := ob.EndTime.Add(30 * time.Minute)
					if (obStart.Before(tp) || obStart.Equal(tp)) && tp.Before(obEndBuf) {
						occupiedAtTP += ob.PartySize
					}
				}
				if occupiedAtTP+booking.PartySize > table.Capacity {
					c.JSON(http.StatusConflict, gin.H{
						"error": "Cannot check in: total party size would exceed table capacity during this dining window.",
					})
					return
				}
			}
		} else {
			if booking.PartySize > table.Capacity {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot check in: party size exceeds table capacity."})
				return
			}
		}

		// Update booking status to checked_in
		update := bson.M{
			"$set": bson.M{
				"status":     "checked_in",
				"updated_at": time.Now(),
			},
		}
		_, err = getBookingCollection().UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check in booking"})
			return
		}

		// Update table status in DB
		updateTableStatusForCurrentTime(ctx, booking.TableID)

		c.JSON(http.StatusOK, gin.H{"message": "Booking checked in successfully"})
	}
}

// @Summary Cancel Booking
// @Description Cancel a booking. Customers can cancel their own, staff/admin can cancel any.
// @Tags Booking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Success 200 {object} models.SuccessResponse "Booking cancelled"
// @Failure 400 {object} models.ErrorResponse "Invalid ID or booking already cancelled"
// @Failure 403 {object} models.ErrorResponse "Access denied"
// @Failure 404 {object} models.ErrorResponse "Booking not found"
// @Router /bookings/{id}/cancel [put]
func CancelBooking() gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(bookingID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
			return
		}

		var booking models.Booking
		err = getBookingCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&booking)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		// Customers can only cancel their own bookings
		userType := c.GetString("user_type")
		if userType == "USER" && booking.UserEmail != c.GetString("email") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		if booking.Status == "cancelled" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Booking is already cancelled"})
			return
		}

		// Cancel the booking
		update := bson.M{
			"$set": bson.M{
				"status":     "cancelled",
				"updated_at": time.Now(),
			},
		}
		_, err = getBookingCollection().UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
			return
		}

		// Recalculate table status based on remaining active bookings
		updateTableStatusForCurrentTime(ctx, booking.TableID)

		c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
	}
}

// @Summary Update Minimum Entry Time
// @Description Staff/admin shifts the minimum entry time (cutoff time) of a booking
// @Tags Booking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Param body body models.Booking true "New minimum entry time"
// @Success 200 {object} models.SuccessResponse "Minimum entry time updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 404 {object} models.ErrorResponse "Booking not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /bookings/{id}/min-entry-time [put]
func UpdateMinEntryTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(bookingID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
			return
		}

		var input struct {
			MinEntryTime time.Time `json:"min_entry_time"`
		}
		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var booking models.Booking
		err = getBookingCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&booking)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}

		if booking.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending bookings can have their entry cutoff time shifted"})
			return
		}

		// Validation rules:
		// 1. Must not be less than 30 minutes after StartTime
		minLimit := booking.StartTime.Add(30 * time.Minute)
		if input.MinEntryTime.Before(minLimit) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum entry time cannot be less than 30 minutes after the booking start time"})
			return
		}

		// 2. Must not be less than 30 minutes before EndTime
		maxLimit := booking.EndTime.Add(-30 * time.Minute)
		if input.MinEntryTime.After(maxLimit) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum entry time cannot be less than 30 minutes before the booking end time"})
			return
		}

		// Update the min_entry_time
		update := bson.M{
			"$set": bson.M{
				"min_entry_time": input.MinEntryTime,
				"updated_at":     time.Now(),
			},
		}
		_, err = getBookingCollection().UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update minimum entry time"})
			return
		}

		// Recalculate table status
		updateTableStatusForCurrentTime(ctx, booking.TableID)

		c.JSON(http.StatusOK, gin.H{"message": "Minimum entry time updated successfully"})
	}
}

// StartBookingAutoCanceller starts a background goroutine to automatically cancel pending bookings that have missed their cutoff time, and auto-complete checked-in bookings that have finished.
func StartBookingAutoCanceller(ctx context.Context) {
	go func() {
		// Run once immediately on start in background
		fmt.Println("⏳ [Auto-Worker] Initializing background booking auto-canceller...")
		runAutoCancellation(ctx)

		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				runAutoCancellation(ctx)
			case <-ctx.Done():
				fmt.Println("🛑 [Auto-Worker] Booking auto-canceller stopped.")
				return
			}
		}
	}()
}

func runAutoCancellation(ctx context.Context) {
	fmt.Println("⏳ [Auto-Worker] Starting auto-cancellation and auto-completion check...")
	bookingCol := getBookingCollection()
	now := time.Now()

	// 1. Auto-cancel pending bookings whose cutoff time has passed
	cancelFilter := bson.M{
		"status": "pending",
	}

	cancelCursor, err := bookingCol.Find(ctx, cancelFilter)
	if err != nil {
		fmt.Printf("❌ [Auto-Worker] Error finding pending bookings: %v\n", err)
	} else {
		var pendingBookings []models.Booking
		if err := cancelCursor.All(ctx, &pendingBookings); err != nil {
			fmt.Printf("❌ [Auto-Worker] Error decoding pending bookings: %v\n", err)
		} else {
			fmt.Printf("⏳ [Auto-Worker] Found %d pending bookings to evaluate.\n", len(pendingBookings))
			for _, b := range pendingBookings {
				// Calculate cutoff time (use stored value or compute fallback for legacy documents)
				var cutoff time.Time
				if b.MinEntryTime.IsZero() {
					durationHours := b.EndTime.Sub(b.StartTime).Hours()
					limitMinutes := int(durationHours) * 30
					if limitMinutes < 30 {
						limitMinutes = 30
					}
					cutoff = b.StartTime.Add(time.Duration(limitMinutes) * time.Minute)
				} else {
					cutoff = b.MinEntryTime
				}

				if now.After(cutoff) {
					fmt.Printf("⏳ [Auto-Worker] Pending booking %s is expired (cutoff: %s, now: %s). Cancelling...\n", b.ID.Hex(), cutoff.Format("15:04:05"), now.Format("15:04:05"))
					update := bson.M{
						"$set": bson.M{
							"status":         "cancelled",
							"min_entry_time": cutoff, // populate it so it's not zero anymore
							"updated_at":     now,
						},
					}
					_, err := bookingCol.UpdateOne(ctx, bson.M{"_id": b.ID}, update)
					if err != nil {
						fmt.Printf("❌ [Auto-Worker] Error updating booking %s to cancelled: %v\n", b.ID.Hex(), err)
					} else {
						fmt.Printf("❌ [Auto-Worker] Booking %s for Table %s auto-cancelled due to no-show.\n", b.ID.Hex(), b.TableID)
					}
				}
			}
		}
		cancelCursor.Close(ctx)
	}

	// 2. Auto-complete checked-in bookings whose end time has passed
	completeFilter := bson.M{
		"status":   "checked_in",
		"end_time": bson.M{"$lt": now},
	}

	completeCursor, err := bookingCol.Find(ctx, completeFilter)
	if err != nil {
		fmt.Printf("❌ [Auto-Worker] Error finding checked-in bookings: %v\n", err)
	} else {
		var finishedBookings []models.Booking
		if err := completeCursor.All(ctx, &finishedBookings); err != nil {
			fmt.Printf("❌ [Auto-Worker] Error decoding checked-in bookings: %v\n", err)
		} else {
			fmt.Printf("⏳ [Auto-Worker] Found %d finished checked-in bookings to complete.\n", len(finishedBookings))
			for _, b := range finishedBookings {
				update := bson.M{
					"$set": bson.M{
						"status":     "completed",
						"updated_at": now,
					},
				}
				_, err := bookingCol.UpdateOne(ctx, bson.M{"_id": b.ID}, update)
				if err != nil {
					fmt.Printf("❌ [Auto-Worker] Error updating booking %s to completed: %v\n", b.ID.Hex(), err)
				} else {
					fmt.Printf("✅ [Auto-Worker] Booking %s for Table %s auto-completed (dining time finished).\n", b.ID.Hex(), b.TableID)
				}
			}
		}
		completeCursor.Close(ctx)
	}

	// 3. Update status of all tables to guarantee database consistency
	tableCol := getTableCollection()
	cursor, err := tableCol.Find(ctx, bson.M{})
	if err != nil {
		fmt.Printf("❌ [Auto-Worker] Error finding tables for status sync: %v\n", err)
	} else {
		var tables []models.Table
		if err := cursor.All(ctx, &tables); err != nil {
			fmt.Printf("❌ [Auto-Worker] Error decoding tables for status sync: %v\n", err)
		} else {
			fmt.Printf("⏳ [Auto-Worker] Syncing status for %d tables...\n", len(tables))
			for _, t := range tables {
				updateTableStatusForCurrentTime(ctx, t.ID.Hex())
			}
			fmt.Println("✅ [Auto-Worker] Table status sync completed.")
		}
		cursor.Close(ctx)
	}
}


