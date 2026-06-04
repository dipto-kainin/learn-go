package routes

import (
	"restroBackend/controllers"
	"restroBackend/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	router.POST("/auth/signup", controllers.Signup())
	router.POST("/auth/login", controllers.Login())
	router.POST("/auth/refresh", controllers.Refresh())
	router.GET("/auth/user", middleware.Authentication(), controllers.GetUser())
}
