package authentication

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("my_secret_key")

// GenerateToken generates a JWT for a given username
func GenerateToken(username string) (string, error) {
	fmt.Println("GenerateToken")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})
	return token.SignedString(jwtKey)
}

// ParseToken validates a JWT and extracts claims
func parseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	
	claims, _ := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	return username, nil
}

// Middleware to authenticate users
func AuthMiddleware(c *gin.Context) {
	fmt.Println("AuthMiddleware")
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		c.Abort()
		return
	}

	username, err1 := parseToken(tokenString)

	if err1 != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"}) //FIXME
		c.Abort()
		return
	}

	c.Set("username", username)
	c.Next()
}
