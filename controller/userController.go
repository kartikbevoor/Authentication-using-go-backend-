package controller

import (
	"Authentication_Using_JWT/database"
	"Authentication_Using_JWT/helper"
	"Authentication_Using_JWT/models"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword()

func VerifyPassword()

func Signup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := userCollection.CountDocuments(c, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking the email"})
			return
		}

		count, err = userCollection.CountDocuments(c, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking phone number"})
			return
		}

		if count > 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone already exists"})
			return
		}
	}
}

func Login()

func GetUsers()

// this function returns another function that will handle the HTTP request.
func GetUser() gin.HandlerFunc { // this function returns a Gin handler function.
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id") // get userid from url

		// with the function MatchUserTypeToUid we are trying to check, if the user is of admin or normal user
		// if normal user he can only access his info
		// userId: this stores the id of the whom we want to access
		if err := helper.MatchUserTypeToUid(ctx, userId); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		// creation of context: to prevent the database operations from hanging forever,
		// automatically cancel long-running operations
		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		// Query MongoDB: This line does the database lookup.
		// FindOne: Search for a single document.
		err := userCollection.FindOne(c, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, user)
	}
}
