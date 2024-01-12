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


	wg.Add(1)

	// Run the HTTP server in a goroutine
	go func() {
		defer wg.Done()

		serverConfig := controller.ServerConfig{
			Collection: collection,
			Port:       ":7082",
		}
		controller.StartServer(serverConfig)
	}()

	
	// Run the Discord server in a goroutine
	// go func() {
	// 	defer wg.Done()
	// 	controller.StartDiscord(collection)
	// }()


	// go func() {
	// 	defer wg.Done()
	// 	controller.StartTelegram(collection)
	// }()

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