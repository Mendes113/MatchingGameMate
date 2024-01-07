package main

import (
	"context"
	"gameMatcher/controller"
	"gameMatcher/model"
	"gameMatcher/service"
	"log"
	"sync"
)

func main() {
	mongoClient := model.MongoConnection()
	defer mongoClient.Disconnect(context.Background())

	collection := mongoClient.Database("gamestore").Collection("games")
	if collection == nil {
		log.Fatal("Collection is nil")
	}
	
	var wg sync.WaitGroup

	// gamesResponse, err := service.GetAllGames()
    // if err != nil {
    //     log.Fatal(err)
    // }

    // // Convert []service.Game to []model.Game
    // var games []model.Game
    // for _, game := range gamesResponse.Results {
    //     // Assuming model.Game has the same structure as service.Game
    //     convertedGame := model.Game{
    //         Name:   game.Name,
    //         Rating: game.Rating,
    //         Genres: convertGenres(game.Genres),
    //     }
    //     games = append(games, convertedGame)
    // }

    // model.SaveGamesInMongoDB(games, collection)

	// Increment the WaitGroup counter for each server
	wg.Add(2)

	// Run the HTTP server in a goroutine
	go func() {
		defer wg.Done()
		controller.StartServer(collection)
	}()

	// Run the Discord server in a goroutine
	go func() {
		defer wg.Done()
		controller.StartDiscord(collection)
	}()


	go func() {
		defer wg.Done()
		controller.StartTelegram(collection)
	}()

	// Wait for both servers to finish
	wg.Wait()
}


func convertGenres(serviceGenres []service.Genre) []model.Genre {
    var modelGenres []model.Genre
    for _, sg := range serviceGenres {
        mg := model.Genre{Name: sg.Name}
        modelGenres = append(modelGenres, mg)
    }
    return modelGenres
}