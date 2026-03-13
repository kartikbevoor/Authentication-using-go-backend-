package routes

import (
	"Authentication_Using_JWT/controller"
	"Authentication_Using_JWT/middleware"

	"github.com/gin-gonic/gin"
)

// function used to register routes related to users.
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate()) // middleware
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
}
