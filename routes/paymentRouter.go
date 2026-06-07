package routes

import (
	"restroBackend/controllers"
	"restroBackend/middleware"

	"github.com/gin-gonic/gin"
)

func PaymentRoutes(router *gin.Engine) {
	router.POST("/payments/create-order", middleware.Authentication(), controllers.CreateRazorpayOrder())
	router.POST("/payments/verify", middleware.Authentication(), controllers.VerifyRazorpayPayment())
	router.POST("/payments/request-cash", middleware.Authentication(), controllers.RequestCashPayment())
	router.POST("/payments/confirm-cash", middleware.Authentication(), controllers.ConfirmCashPayment())
}
