package routes

import (
	"basic-backend/controllers"
	"basic-backend/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	router.POST("/auth/signup", controllers.Signup())
	router.POST("/auth/login", controllers.Login())
	router.GET("/auth/user", middleware.Authentication(), controllers.GetUser())
}
