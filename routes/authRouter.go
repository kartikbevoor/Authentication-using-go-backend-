package routes

import (
	"Authentication_Using_JWT/controller"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) { // *gin.Engine → pointer to a Gin router instance	// *gin.Engine represents the main router of the web application.
	incomingRoutes.POST("users/signup", controller.Signup())
	incomingRoutes.POST("users/login", controller.Login())
}
