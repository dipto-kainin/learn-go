package routes

import (
	"restroBackend/controllers"
	"restroBackend/middleware"

	"github.com/gin-gonic/gin"
)

func OrderItemRoutes(router *gin.Engine) {
	// Primary hyphenated routes
	router.GET("/order-items", middleware.Authentication(), controllers.GetOrderItems())
	router.GET("/order-items/:id", middleware.Authentication(), controllers.GetOrderItem())
	router.POST("/order-items", middleware.Authentication(), controllers.CreateOrderItem())
	router.PUT("/order-items/:id", middleware.Authentication(), controllers.UpdateOrderItem())
	router.DELETE("/order-items/:id", middleware.Authentication(), controllers.DeleteOrderItem())

	// Non-hyphenated fallback routes
	router.GET("/orderitems", middleware.Authentication(), controllers.GetOrderItems())
	router.GET("/orderitems/:id", middleware.Authentication(), controllers.GetOrderItem())
	router.POST("/orderitems", middleware.Authentication(), controllers.CreateOrderItem())
	router.PUT("/orderitems/:id", middleware.Authentication(), controllers.UpdateOrderItem())
	router.DELETE("/orderitems/:id", middleware.Authentication(), controllers.DeleteOrderItem())
}