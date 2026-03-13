package main

import (
	"Authentication_Using_JWT/routes"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router := gin.New() // creates a new Gin router (engine) and returns a pointer to a Gin Engine.
	// router := gin.Default()	// automatically adds middleware like: logger and recovery
	// gin.New() creates a clean router with no middleware.

	router.Use(gin.Logger()) // adds middleware to the router.
	// | Part           | Meaning                                       |
	// | -------------- | --------------------------------------------- |
	// | `router.Use()` | Adds middleware to router                     |
	// | `gin.Logger()` | Built-in middleware for logging HTTP requests |
	// gin.Logger()	// logs every HTTP request to the console.

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"success": "Access granted for api-1"})
	})

	router.GET("/api-2", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	router.Run(":" + port)
}
