package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/lipezaballa/FaaS-system/api-server/controllers"
	"github.com/lipezaballa/FaaS-system/api-server/natsConnection"
	"github.com/lipezaballa/FaaS-system/reverse-proxy/authentication"
	"github.com/lipezaballa/FaaS-system/shared"
	"github.com/nats-io/nats.go"
)

var natsConn *shared.NatsConnection

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	log.Println("Starting FaaS API Server...")


	//natsURL := os.Getenv("NATS_URL")
	natsURL := "nats://nats_server:4222"
	log.Println("ejecutando testNATS")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Error al conectar a NATS: %v", err)
	}
	defer nc.Close()

	natsConn, err = natsConnection.InitJetStream(nc, "queue.messages.worker")
	if err != nil {
		log.Fatalf("Error al iniciar JetStream y KV Store: %v", err)
	}

	// Initialize router
	router := gin.Default()

	// Public routes
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)
	router.GET("/", initPage)
	router.DELETE("/database", deleteDataBase)

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

func deleteDataBase(c *gin.Context) {
	natsConnection.DeleteAllKeysFromKV()
}