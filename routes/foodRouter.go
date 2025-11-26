package routes

import (
	"basic-backend/controllers"
	"basic-backend/middleware"

	"github.com/gin-gonic/gin"
)

func FoodRoutes(router *gin.Engine) {
	router.GET("/foods", controllers.GetFoods())
	router.GET("/foods/:id", controllers.GetFood())
	router.POST("/foods", middleware.Authentication(), middleware.RequireAdmin(), controllers.CreateFood())
	router.PUT("/foods/:id", middleware.Authentication(), middleware.RequireAdmin(), controllers.UpdateFood())
	router.DELETE("/foods/:id", middleware.Authentication(), middleware.RequireAdmin(), controllers.DeleteFood())
}