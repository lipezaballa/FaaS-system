package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lipezaballa/FaaS-system/authentication"
)

var functions = map[string]map[string]string{} // username:function_name:image_reference

func CreateMapForUser(req *authentication.Request) {
	functions[req.Username] = make(map[string]string)
}

// Register a new function
func RegisterFunction(c *gin.Context) {
	fmt.Println("RegisterFunction")
	type Request struct {
		Name        string `json:"name"`
		ImageRef    string `json:"image_ref"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	username := c.GetString("username")
	if _, exists := functions[username][req.Name]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Function already exists"})
		return
	}

	functions[username][req.Name] = req.ImageRef
	c.JSON(http.StatusCreated, gin.H{"message": "Function registered successfully"})
}

// Delete a function
func DeleteFunction(c *gin.Context) {
	fmt.Println("DeleteFunction")
	functionName := c.Param("function_name")
	username := c.GetString("username")

	if _, exists := functions[username][functionName]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		return
	}

	delete(functions[username], functionName)
	c.JSON(http.StatusOK, gin.H{"message": "Function deleted successfully"})
}

// Invoke a function
func InvokeFunction(c *gin.Context) {
	fmt.Println("InvokeFunction")
	functionName := c.Param("function_name")
	username := c.GetString("username")

	imageRef, exists := functions[username][functionName]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		return
	}

	type Request struct {
		Param string `json:"param"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Simulate function execution
	result := "Executed " + functionName + " with param: " + req.Param
	log.Printf("Running container with image: %s", imageRef)
	c.JSON(http.StatusOK, gin.H{"result": result})
}
