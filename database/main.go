package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/casbin/casbin/v2/persist"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client Database instance
var Client *mongo.Client

// Adapter for casbin
var Adapter persist.BatchAdapter

// DBinstance func
func DBinstance() *mongo.Client {
	fmt.Println(os.Getenv("MONGODB_URI"))
	cmdMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			log.Print(evt.Command)
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")).SetMonitor(cmdMonitor))
	// client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}

	defer cancel()
	fmt.Println("Connected to MongoDB!")

	return client
}

// Connect to MongoDB
func Connect() {
	Client = DBinstance()
}

// OpenCollection is a  function makes a connection with a collection in the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database(os.Getenv("MONGODB_DB_NAME")).Collection(collectionName)
	return collection
}
