package main

import (
	"context"
	"gameMatcher/controller"
	"gameMatcher/model"
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

	// Use a WaitGroup to wait for both servers to finish
	var wg sync.WaitGroup

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

	// Wait for both servers to finish
	wg.Wait()
}
