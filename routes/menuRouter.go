package routes

import (
	"basic-backend/controllers"
	"basic-backend/middleware"

	"github.com/gin-gonic/gin"
)

func MenuRoutes(router *gin.Engine) {
	router.GET("/menus", controllers.GetMenus())
	router.GET("/menus/:id", controllers.GetMenu())
	router.POST("/menus", middleware.Authentication(), middleware.RequireAdmin(), controllers.CreateMenu())
	router.PUT("/menus/:id", middleware.Authentication(), middleware.RequireAdmin(), controllers.UpdateMenu())
	router.DELETE("/menus/:id", middleware.Authentication(), middleware.RequireAdmin(), controllers.DeleteMenu())
}