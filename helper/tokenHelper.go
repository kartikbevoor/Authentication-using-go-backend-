package helper

import (
	"Authentication_Using_JWT/database"
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	UId       string
	UserType  string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

// Function to generate JWT access and refresh tokens
func GenerateAllTokens(email, firstName, lastName, userType, uId string) (string, string, error) {

	// Access token claims (payload data)
	claims := &SignedDetails{ // custom claims
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UId:       uId,
		UserType:  userType,
		StandardClaims: jwt.StandardClaims{ // StandardClaims are predefined JWT fields defined by JSON Web Token.
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // current time + 24 hours
		},
	}

	// Refresh token claims
	refreshClaims := &SignedDetails{
		UId: uId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		},
	}

	// Generate access token
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	// .SignedString(...) : is called on the result of .NewWithClaims
	// jwt.NewWithClaims : func NewWithClaims(method SigningMethod, claims Claims) *Token
	// *Token : A pointer to a JWT Token struct.
	// .SignedString([]byte(SECRET_KEY)) : func (t *Token) SignedString(key interface{}) (string, error) // This is a method on the Token struct.
	// | Parameter | Type          | Meaning                     |
	// | --------- | ------------- | --------------------------- |
	// | `key`     | `interface{}` | secret key used for signing |

	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

// | Parameter                | Meaning           |
// | ------------------------ | ----------------- |
// | `jwt.SigningMethodHS256` | signing algorithm |
// | `claims`                 | payload data      |

// signing method : HS256
// SignedString([]byte(SECRET_KEY)) : cryptographically signs the token using the secret key

// JWT flow
// user login : server generates access tokens and referesh tokens.
// user : uses, access tokens to access the data and make contacts with the server.
// when access token expires : user uses referesh tokens to get access tokens.

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	updateObj := primitive.D{
		{Key: "token", Value: signedToken},
		{Key: "refreshToken", Value: signedRefreshToken},
		{Key: "updated_at", Value: time.Now()},
	}

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: updateObj},
		},
		&opt,
	)

	if err != nil {
		log.Println("Failed to update tokens:", err)
		return err
	}

	return nil
}
