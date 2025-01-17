package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lipezaballa/FaaS-system/api-server/natsConnection"
	"github.com/lipezaballa/FaaS-system/reverse-proxy/authentication"
)

// Login user and return a token
func LoginUser(c *gin.Context) {
	fmt.Println("LoginUser")
	var req authentication.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	storedPasswordEntry, exists := natsConnection.GetValue(req.Username)
	if (!exists) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User no exists"})
		return
	}
	storedPassword := string(storedPasswordEntry.Value())
	printVariables(req.Username, req.Password, storedPassword) //FIXME: nil in storedPassword?
	if !authentication.CheckPasswordHash(req.Password, storedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}


	tokenString, err := authentication.GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// Register a new user
func RegisterUser(c *gin.Context) {
	fmt.Println("RegisterUser")
	var req authentication.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if _, exists := natsConnection.GetValue(req.Username); exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	hashedPassword, err := authentication.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Error hashing password"})
		return
	}

	natsConnection.StoreUser(req.Username, hashedPassword)
	natsConnection.PrintValues()
	printVariables(req.Username, req.Password, hashedPassword)

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Get user info
func GetUserInfo(c *gin.Context) {
	fmt.Println("GetUser")
	username := c.GetString("username")
	c.JSON(http.StatusOK, gin.H{"username": username})
}

func printVariables(username string, password string, hashedPassword string) {
	fmt.Println("user: ", username)
	fmt.Println("pass: ", password)
	fmt.Println("hashedPass: ", hashedPassword)
}