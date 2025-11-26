package routes

import (
	"basic-backend/controllers"
	"basic-backend/middleware"

	"github.com/gin-gonic/gin"
)

func TableRoutes(router *gin.Engine) {
	router.GET("/tables", controllers.GetTables())
	router.GET("/tables/:id", controllers.GetTable())
	router.POST("/tables", middleware.Authentication(), middleware.RequireAdmin(), controllers.CreateTable())
	router.PUT("/tables/:id", middleware.Authentication(), middleware.RequireAdmin(), controllers.UpdateTable())
	router.DELETE("/tables/:id", middleware.Authentication(), middleware.RequireAdmin(), controllers.DeleteTable())
}