package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gameMatcher/model"
	"log"
	"net/http"
	"github.com/jaytaylor/html2text"
	"github.com/gocolly/colly/v2"
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


func GetSteamGameIdUsingName(collection *mongo.Collection, name string) (SteamGame, error) {
    // Use o contexto de fundo como exemplo, você pode querer usar um contexto apropriado no seu aplicativo
    ctx := context.TODO()
	log.Printf("Buscando jogo com o nome '%s' na coleção", name)
    game, err := model.GetSteamGameIdUsingName(collection, ctx, name)
    if err != nil {
        return SteamGame{}, fmt.Errorf("erro ao buscar jogo: %v", err)
    }
	log.Printf("Jogo encontrado: %v e %v", game, game.AppID)
	
    return SteamGame(game), nil
}

func ScrapReviewsFromSteam(game SteamGame, collection *mongo.Collection) (string, error) {
	gameID, err := GetSteamGameIdUsingName(collection, game.Name)
	if err != nil {
		return "", fmt.Errorf("error getting Steam game ID: %v", err)
	}
	game.AppID = gameID.AppID
	steambaseURL := "https://store.steampowered.com/app/"
	url := fmt.Sprintf("%s%d/%s", steambaseURL, game.AppID, game.Name)
	fmt.Println(url)

	// Create a new collector
	c := colly.NewCollector()

	// Variable to store reviews
	var reviews string

	// Encontre e raspe o seletor desejado
	c.OnHTML(".glance_ctn .user_reviews", func(e *colly.HTMLElement) {
		rawReviews := e.Text
		cleanedReviews, err := html2text.FromString(rawReviews, html2text.Options{PrettyTables: true})
		if err != nil {
			log.Println("Error cleaning reviews:", err)
			reviews = rawReviews // Se houver um erro, use o texto original
		} else {
			reviews = cleanedReviews
		}
		fmt.Println("User Reviews:", reviews)
	})

	// Visite a URL do jogo
	err = c.Visit(url)
	if err != nil {
		log.Fatal(err)
		return "Error Ao visitar URL", err
	}

	return reviews, nil
}


type SteamGameAlias SteamGame

func (s *SteamGameAlias) ToSteamGame() SteamGame {
	return SteamGame(*s)
}

func GetGameReviewFromGameName(collection *mongo.Collection, name string) (string, error) {
	game, err := GetSteamGameIdUsingName(collection, name)
	if err != nil {
		return "", fmt.Errorf("error getting Steam game ID: %v", err)
	}

	reviews, err := ScrapReviewsFromSteam(game, collection)
	if err != nil {
		return "", fmt.Errorf("error getting Steam game reviews: %v", err)
	}

	return reviews, nil
}