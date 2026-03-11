package controller

import (
	"Authentication_Using_JWT/database"
	"Authentication_Using_JWT/helper"
	"Authentication_Using_JWT/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword()

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("email of password is incorrect")
		check = false
	}

	return check, msg
}

func Signup() gin.HandlerFunc { // This function returns a Gin HTTP handler.
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		// c → context used for DB operations, cancel → function that releases resources
		defer cancel()

		var user models.User

		if err := ctx.BindJSON(&user); err != nil { // This converts the incoming JSON request into the user struct.
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		emailCount, err := userCollection.CountDocuments(c, bson.M{"email": user.Email})
		// the above is same as SELECT COUNT(*) FROM users WHERE email="rahul@gmail.com"
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking the email"})
			return
		}

		phoneCount, err := userCollection.CountDocuments(c, bson.M{"phone": user.Phone})
		// the above is to check if any user already exists with same number
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking phone number"})
			return
		}

		if emailCount > 0 || phoneCount > 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone already exists"})
			return
		}

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// Creating MongoDB ObjectID
		user.Id = primitive.NewObjectID() // Creates a unique MongoDB _id.
		// Converting ObjectID to String
		user.UserId = user.Id.Hex() // Converts ObjectID into a hex string.
		// Generating Authentication Tokens
		token, refereshToken, _ := helper.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, *&user.UserId)
		user.Token = &token
		user.RefereshToken = &refereshToken

		// Inserting User into MongoDB
		resultInsertionNumber, insertErr := userCollection.InsertOne(c, user) // This inserts the user document into MongoDB.
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		// defer cancel()
		ctx.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), time.Second*100)
		var user models.User
		var foundUser models.User

		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(c, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Email or passoword is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
	}
}

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
