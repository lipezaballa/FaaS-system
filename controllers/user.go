package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lipezaballa/FaaS-system/authentication"
)

var users = map[string]string{}   // username:password

// Login user and return a token
func LoginUser(c *gin.Context) {
	fmt.Println("LoginUser")
	var req authentication.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	storedPassword, exists := users[req.Username]
	if !exists || storedPassword != req.Password {
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

	if _, exists := users[req.Username]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	users[req.Username] = req.Password
	CreateMapForUser(&req)

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Get user info
func GetUserInfo(c *gin.Context) {
	fmt.Println("GetUser")
	username := c.GetString("username")
	c.JSON(http.StatusOK, gin.H{"username": username})
}