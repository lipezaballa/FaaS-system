package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lipezaballa/FaaS-system/api-server/natsConnection"
)

//var functions = map[string]map[string]string{} // username:function_name:image_reference

/*func CreateMapForUser(req *authentication.Request) {
	functions[req.Username] = make(map[string]string)
}*/

// Register a new function
func RegisterFunction(c *gin.Context) {
	fmt.Println("RegisterFunction")
	

	var req FunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	username := c.GetString("username")
	key := fmt.Sprintf("users/%s/functions/%s", username, req.Name)
	//if _, exists := functions[username][req.Name]; exists {
	if _, exists := natsConnection.GetValue(key); exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Function already exists"})
		return
	}

	//functions[username][req.Name] = req.ImageRef

	natsConnection.StoreFunction(username, req.Name, req.ImageRef)
	natsConnection.PrintValues()
	//printFunctions()
	c.JSON(http.StatusCreated, gin.H{"message": "Function registered successfully"})
}

// Delete a function
func DeleteFunction(c *gin.Context) {
	fmt.Println("DeleteFunction")
	functionName := c.Param("function_name")
	username := c.GetString("username")

	key := fmt.Sprintf("users/%s/functions/%s", username, functionName)
	//imageRef, exists := natsConnection.GetValue(key)
	//if _, exists := functions[username][functionName]; !exists {
	if _, exists := natsConnection.GetValue(key); !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		return
	}

	//delete(functions[username], functionName)

	//key := fmt.Sprintf("users/%s/functions/%s", username, functionName)
	natsConnection.DeleteKeyFromKV(key)
	natsConnection.PrintValues()
	//printFunctions()
	c.JSON(http.StatusOK, gin.H{"message": "Function deleted successfully"})
}

// Invoke a function
func InvokeFunction(c *gin.Context) {
	functionName := c.Param("function_name")
	username := c.GetString("username")

	//imageRef, exists := functions[username][functionName]
	key := fmt.Sprintf("users/%s/functions/%s", username, functionName)
	imageRef, exists := natsConnection.GetValue(key)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		return
	}

	var req FunctionParameter
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	request := fmt.Sprintf("%s|%s|%s", username, string(imageRef.Value()), req.Param)
	fmt.Println("request: ", request)
	resp, err := natsConnection.SendRequest(request)
	if err != nil {
		log.Println("error in sendRequest, ", err)
	}
	// FIXME: Simulate function execution
	result := "Executed " + functionName + " with param: " + req.Param + ", Result: " + string(resp.Data)
	log.Println(result)
	c.JSON(http.StatusOK, gin.H{"result": string(resp.Data)})
}

/*func printFunctions() {
	fmt.Println(functions)
}*/
