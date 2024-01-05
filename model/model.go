package model

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func MongoConnection() *mongo.Client {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = mongoClient.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	return mongoClient
}


func GetCollection(collectionName string) *mongo.Collection {
	mongoClient := MongoConnection()
	collection := mongoClient.Database("gamestore").Collection(collectionName)
	if collection == nil {
		log.Fatal("Collection is nil")
	}
	return collection
}


func GetCollectionFromDB(dbName, collectionName string) *mongo.Collection {
	mongoClient := MongoConnection()
	collection := mongoClient.Database(dbName).Collection(collectionName)
	if collection == nil {
		log.Fatal("Collection is nil")
	}
	return collection
}

func SaveGamesInMongoDB(games []Game, collection *mongo.Collection) error {
	var gamesInterface []interface{}
	for _, game := range games {
		gamesInterface = append(gamesInterface, game)
	}
	_, err := collection.InsertMany(context.Background(), gamesInterface)
	if err != nil {
		return err
	}
	return nil
}

type Game struct {
	Name    string   `json:"name"`
	Genres  []Genre  `json:"genres"`
	Rating  float64  `json:"rating"`
  }
  
  type Genre struct {
	Name string `json:"name"`
  }

type GameListResponse struct {
	Results []Game `json:"results"`
	Next    string  `json:"next"`
}
