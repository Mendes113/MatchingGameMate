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



		// games, err := getAllGames()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// for _, game := range games.Results {
		// 	_, err := collection.InsertOne(context.Background(), game)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// }

		

		// for _, game := range gamesFromDB {
		// 	fmt.Println(game.Name)
		// }

		
		

		// user1 := UserChoices{
		// 	Username: "user1",
		// 	Genres:   []string{"Action", "Adventure", "RPG"},}
		// user2 := UserChoices{
		// 	Username: "user2",
		// 	Genres:   []string{"Action", "Adventure", "Strategy"},}
			
		// equalGames, err := equalGamesIn2Users(collection, user1, user2)
		// if err != nil {
		// 	log.Fatal(err)
		// }



		// //best games for the 2 users
		// top5Games, err := top5GamesRated(equalGames)
		// if err != nil {
		// 	log.Fatal(err)
		// }


		// topGamesUserOne , err := gamesForOneUser(collection, user1)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Println("Top games for user 1:")
		// for _, game := range topGamesUserOne {
		// 	fmt.Println(game.Name)
		// }

		

		// fmt.Println("Top 5 games:")
		// for _, game := range top5Games {
		// 	fmt.Println(game.Name)
		// }


		// for  _, game := range top5Games {
		// 	similarGames, err := getSimilarGames(game, "https://gameslikefinder.com/")
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	fmt.Println(similarGames)
		// }

		




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



