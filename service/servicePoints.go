package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gameMatcher/model"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)
type SteamGame struct {
	AppID int `json:"appid"`
	Name string   `json:"name"`
}
func GetGamesFromSteamAPI(collection *mongo.Collection) (int, error) {
	// Uncomment the following lines if using Steam API key
	// apiKey := os.Getenv("STEAM_KEY")
	// if apiKey == "" {
	// 	return 0, fmt.Errorf("a chave da API da Steam não está configurada")
	// }

	url := "https://api.steampowered.com/ISteamApps/GetAppList/v2/"

	response, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("erro ao obter a lista de aplicativos: %v", err)
	}
	defer response.Body.Close()

	var appList struct {
		Apps []SteamGame `json:"apps"`
	}
	err = json.NewDecoder(response.Body).Decode(&appList)
	if err != nil {
		return 0, fmt.Errorf("erro ao decodificar a resposta JSON: %v", err)
	}

	var convertedGames []model.SteamGame
	for _, game := range appList.Apps {
		convertedGames = append(convertedGames, model.SteamGame{
			AppID: game.AppID,
			Name:  game.Name,
		})
	}

	err = model.SaveSteamResponse(collection, convertedGames)
	if err != nil {
		return 0, fmt.Errorf("erro ao salvar na coleção: %v", err)
	}

	return len(convertedGames), nil
}


func GetGamesFromSteamAPI1(collection *mongo.Collection) ([]SteamGame, error) {
	// Uncomment the following lines if using Steam API key
	// apiKey := os.Getenv("STEAM_KEY")
	// if apiKey == "" {
	// 	return nil, fmt.Errorf("a chave da API da Steam não está configurada")
	// }

	url := "https://api.steampowered.com/ISteamApps/GetAppList/v2/"

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from Steam API: %v", err)
	}
	defer response.Body.Close()

	var result map[string]struct {
		Apps []SteamGame `json:"apps"`
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	// Assuming you want to insert the data into MongoDB, you can iterate over the games and insert them into the collection
	for _, game := range result["applist"].Apps {
		_, err := collection.InsertOne(context.Background(), bson.M{"name": game.Name, "appid": game.AppID})
		if err != nil {
			// Handle the error or log it
			fmt.Printf("Failed to insert game into MongoDB: %v\n", err)
		}else {
			// Log success
			log.Printf("Game inserted into MongoDB: %s (AppID: %d)\n", game.Name, game.AppID)
		}
	}

	return result["applist"].Apps, nil
}