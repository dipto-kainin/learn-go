package routes

import (
	"basic-backend/controllers"
	"basic-backend/middleware"

	"github.com/gin-gonic/gin"
)

func OrderRoutes(router *gin.Engine) {
	router.GET("/orders", middleware.Authentication(), controllers.GetOrders())
	router.GET("/orders/:id", middleware.Authentication(), controllers.GetOrder())
	router.POST("/orders", middleware.Authentication(), controllers.CreateOrder())
	router.PUT("/orders/:id", middleware.Authentication(), controllers.UpdateOrder())
	router.DELETE("/orders/:id", middleware.Authentication(), controllers.DeleteOrder())
}