package routes

import (
	"basic-backend/controllers"
	"basic-backend/middleware"

	"github.com/gin-gonic/gin"
)

func InvoiceRoutes(router *gin.Engine) {
	router.GET("/invoices", middleware.Authentication(), controllers.GetInvoices())
	router.GET("/invoices/:id", middleware.Authentication(), controllers.GetInvoice())
	router.POST("/invoices", middleware.Authentication(), controllers.CreateInvoice())
	router.PUT("/invoices/:id", middleware.Authentication(), middleware.RequireAdmin(), controllers.UpdateInvoice())
}