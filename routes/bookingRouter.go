package routes

import (
	"restroBackend/controllers"
	"restroBackend/middleware"

	"github.com/gin-gonic/gin"
)

func BookingRoutes(router *gin.Engine) {
	router.POST("/bookings", middleware.Authentication(), controllers.CreateBooking())
	router.GET("/bookings", middleware.Authentication(), controllers.GetBookings())
	router.GET("/bookings/:id", middleware.Authentication(), controllers.GetBooking())
	router.PUT("/bookings/:id/check-in", middleware.Authentication(), middleware.RequireStaffOrAdmin(), controllers.CheckInBooking())
	router.PUT("/bookings/:id/cancel", middleware.Authentication(), controllers.CancelBooking())
	router.PUT("/bookings/:id/min-entry-time", middleware.Authentication(), middleware.RequireStaffOrAdmin(), controllers.UpdateMinEntryTime())
}
