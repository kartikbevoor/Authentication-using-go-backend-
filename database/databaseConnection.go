package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client { // The function returns a pointer to a MongoDB client.
	// A mongo.Client is the object used to: connect to MongoDB, access databases
	// run queries, insert or retrieve documents

	// Loading Environment Variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	MongoDb := os.Getenv("MONGODB_URL") // this reads env variable MONGODB_URL into MongoDb

	// Creating a MongoDB Client
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDb)) // This creates a new MongoDB client instance.
	// mongo.NewClient() → creates a MongoDB client, options.Client() → configures the client, ApplyURI(MongoDb) → sets the connection string

	if err != nil {
		log.Fatal(err)
	}

	// Creating a Context with Timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // This creates a context with a timeout.
	// context.Background() returns a base (root) context.
	// Characteristics: It never gets canceled, It has no deadline, It carries no values
	// Typically used at the top level of programs, like in main() or server request handlers.

	// context.WithTimeout(parent, duration) creates a child context that automatically cancels after a specified time.
	// function signature: func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
	// Context (ctx) – the derived context with a timeout.
	// Cancel function (cancel) – used to manually cancel the context.

	defer cancel() // A function you must call to release resources.

	// Connecting to MongoDB
	err = client.Connect(ctx) // This actually opens the connection to MongoDB using the context.
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB")

	return client // The function returns the connected MongoDB client instance, which is used by the rest of the application
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("cluster0").Collection(collectionName)
	return collection
}
