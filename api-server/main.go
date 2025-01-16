package main

import (
	"fmt"
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

	natsConn, err = natsConnection.InitJetStream(nc, "queue.messages")
	if err != nil {
		log.Fatalf("Error al iniciar JetStream y KV Store: %v", err)
	}

	// Initialize router
	router := gin.Default()

	// Public routes
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)
	router.GET("/", initPage)

	//Store natsConn in Gin Context
	/*router.Use(func(c *gin.Context) { //FIXME, could be used in protected paths?
		// Pasar natsConn a todas las rutas usando Set
		c.Set("natsConn", natsConn)
		c.Next() // Continuar con la siguiente ruta
	})*/

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
	/*natsConn, err := getNatsConnFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}*/
	message := fmt.Sprintf("Peticion 1")
	resp, err := natsConnection.SendRequest(natsConn, message)
	if err != nil {
		log.Println("error in sendRequest, ", err)
	}
	natsConnection.StoreInKv(natsConn, string(resp.Data))
	natsConnection.GetValues(natsConn)
	c.JSON(http.StatusOK, gin.H{"message": "Bienvenido a la API de FaaS-system"})
}

/*func getNatsConnFromContext(c *gin.Context) (*shared.NatsConnection, error) {
	natsConn, exists := c.Get("natsConn")
	if !exists {
		// Si no se encuentra natsConn, retornar un error
		err := errors.New("Conexión a NATS no disponible")
		//c.JSON(http.StatusInternalServerError, gin.H{"error": "Conexión a NATS no disponible"})
		return nil, err
	}

	// Convertir natsConn a su tipo real (*natsConnection.NatsConnection)
	conn, ok := natsConn.(*shared.NatsConnection)
	if !ok {
		// Si la conversión falla, retornar un error
		err := errors.New("Error al convertir la conexión NATS")
		//c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al convertir la conexión NATS"})
		return nil, err
	}
	return conn, nil
}*/