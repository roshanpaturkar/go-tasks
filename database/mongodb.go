package database

import (
	"os"
	"log"
	"time"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoClient() *mongo.Database {
	log.Println("Connecting to MongoDB...")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(os.Getenv("MONGODB_SRV_RECORD")).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")

	return client.Database(os.Getenv("MONGODB_DATABASE"))
}