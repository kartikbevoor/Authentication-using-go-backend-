package controller

import (
	"Authentication_Using_JWT/database"
	"Authentication_Using_JWT/helper"
	"Authentication_Using_JWT/models"
	"context"
	"log"
	"net/http"
	"strconv"
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

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

// verifies whether a user-entered password matches a stored hashed password
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))

	check := true
	msg := ""

	if err != nil {
		msg = "email or password is incorrect"
		check = false
	}

	return check, msg
}

func Signup() gin.HandlerFunc { // This function returns a Gin HTTP handler.
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		// function signature: func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
		// Context: a derived function from a parent, which automatically gets canceled after a specified timeout
		// CancelFunc: A function you should call to manually cancel the context and release resources.
		// c → context used for DB operations, cancel → function that releases resources
		defer cancel()

		var user models.User // create a variable of type user

		if err := ctx.BindJSON(&user); err != nil { // This converts the incoming JSON request into the user struct.
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		emailCount, err := userCollection.CountDocuments(c, bson.M{"email": user.Email}) // How many users exist with the same email address
		// the above is same as SELECT COUNT(*) FROM users WHERE email="rahul@gmail.com"
		// CountDocuments counts the number of documents in a MongoDB collection that match a filter.

		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking the email"})
			return
		}

		hashedPassword := HashPassword(*user.Password)
		user.Password = &hashedPassword

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

		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		// Creating MongoDB ObjectID
		user.Id = primitive.NewObjectID() // Creates a unique MongoDB _id.

		// Converting ObjectID to String
		user.UserId = user.Id.Hex() // Converts ObjectID into a hex string.

		// Generating Authentication Tokens
		token, RefreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, user.UserId)
		user.Token = &token
		user.RefreshToken = &RefreshToken

		// Inserting User into MongoDB
		resultInsertionNumber, insertErr := userCollection.InsertOne(c, user) // This inserts the user document into MongoDB.
		if insertErr != nil {
			// msg := fmt.Sprintf("User item was not created")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User item was not created"})
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
		defer cancel()

		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(c, bson.M{"email": user.Email}).Decode(&foundUser)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Email or passoword is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		if !passwordIsValid {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.UserType, foundUser.UserId)
		helper.UpdateAllTokens(token, refreshToken, foundUser.UserId)
		err = userCollection.FindOne(c, bson.M{"user_id": foundUser.UserId}).Decode(&foundUser)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		ctx.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		if err := helper.CheckUserType(ctx, "ADMIN"); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		matchStage := bson.D{
			{Key: "$match", Value: bson.D{}},
		}

		groupStage := bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "total_count", Value: bson.D{
					{Key: "$sum", Value: 1},
				}},
				{Key: "data", Value: bson.D{
					{Key: "$push", Value: "$$ROOT"},
				}},
			}},
		}

		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
				}},
			}},
		}

		result, err := userCollection.Aggregate(c, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage,
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while fetching users"})
			return
		}

		var allUsers []bson.M
		if err = result.All(c, &allUsers); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(allUsers) == 0 {
			ctx.JSON(http.StatusOK, gin.H{
				"total_count": 0,
				"user_items":  []interface{}{},
			})
			return
		}

		ctx.JSON(http.StatusOK, allUsers[0])
	}
}

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
