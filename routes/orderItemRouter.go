package routes

import (
	"basic-backend/controllers"
	"basic-backend/middleware"

	"github.com/gin-gonic/gin"
)

func OrderItemRoutes(router *gin.Engine) {
	router.GET("/order-items", middleware.Authentication(), controllers.GetOrderItems())
	router.GET("/order-items/:id", middleware.Authentication(), controllers.GetOrderItem())
	router.POST("/order-items", middleware.Authentication(), controllers.CreateOrderItem())
	router.PUT("/order-items/:id", middleware.Authentication(), controllers.UpdateOrderItem())
	router.DELETE("/order-items/:id", middleware.Authentication(), controllers.DeleteOrderItem())
}