package main

import (
	"context"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func loginMongo(username string, password string) error {
	// Create client
	mongoURI := "mongodb+srv://<db_user>:<db_password>@pieselaborator.snqtk.mongodb.net/"
	mongoURI = strings.ReplaceAll(mongoURI, "<db_password>", password)
	mongoURI = strings.ReplaceAll(mongoURI, "<db_user>", username)
	mongoServerAPI := options.ServerAPI(options.ServerAPIVersion1)
	mongoOpts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(mongoServerAPI)
	log.Println("Created client")

	// Connect client
	var err error
	client, err = mongo.Connect(context.TODO(), mongoOpts)
	if err != nil {
		return err
	}
	log.Println("Connected client")

	// Check if connection successfull
	if err := client.Ping(context.TODO(), nil); err != nil {
		return err
	}

	// Retrieve collections
	collection = client.Database("data").Collection("parts")

	return nil
}
