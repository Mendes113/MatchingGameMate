package main

import (
	"context"
	"fmt"
	"gameMatcher/controller"
	"log"

	_ "github.com/chromedp/cdproto/cdp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	
)


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


type UserChoices struct {
	Username string   `json:"username"`
	Genres   []string `json:"genres"`
}



func main() {


		mongoClient := mongoConnection()
		defer mongoClient.Disconnect(context.Background())

		collection := mongoClient.Database("gamestore").Collection("games")
		if collection == nil {
			log.Fatal("Collection is nil")
		}

		controller.StartServer(collection)

	




} 





func mongoConnection() *mongo.Client {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	return client
}



