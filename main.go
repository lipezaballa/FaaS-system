package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/lipezaballa/FaaS-system/authentication"
	"github.com/lipezaballa/FaaS-system/controllers"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	log.Println("Starting FaaS API Server...")

	// Initialize router
	router := gin.Default()

	// Public routes
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)
	router.GET("/", initPage)

	// Protected routes
	protected := router.Group("/")
	protected.Use(authentication.AuthMiddleware)
	{
		protected.POST("/functions", controllers.RegisterFunction)
		protected.DELETE("/functions/:function_name", controllers.DeleteFunction)
		protected.POST("/functions/:function_name/invoke", controllers.InvokeFunction)
		protected.GET("/me", controllers.GetUserInfo)
	}

	// Start server
	router.Run(":8080")
}

func initPage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Bienvenido a la API de FaaS-system"})
}