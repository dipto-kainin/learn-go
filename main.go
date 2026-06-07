package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"restroBackend/controllers"
	"restroBackend/database"
	_ "restroBackend/docs" // Import generated docs
	"restroBackend/middleware"
	"restroBackend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Restaurant Management API
// @version 1.0
// @description This is a restaurant management system API with user authentication, food ordering, and invoice management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@restaurant.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name token

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to MongoDB
	database.ConnectDB()

	// Start background booking auto-cancellation worker
	controllers.StartBookingAutoCanceller(context.Background())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.Default()
	router.Use(middleware.RequestID())
	router.Use(middleware.ResponseWrapper())
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Restaurant API is running",
		})
	})

	// Setup routes
	routes.UserRoutes(router)
	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.OrderRoutes(router)
	routes.TableRoutes(router)
	routes.InvoiceRoutes(router)
	routes.BookingRoutes(router)
	routes.PaymentRoutes(router)

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// will render html file
	router.GET("/", func(c *gin.Context) {
		c.File("./docs/index.html")
	})

	fmt.Printf("🚀 Server starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}