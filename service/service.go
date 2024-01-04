package service

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"fmt"
	"log"

	"time"

	"github.com/chromedp/cdproto/cdp"
	_ "github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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


const apiKey = "3525e0e6ab9a480dbb58207514234680"


func EqualGamesIn2Users(collection *mongo.Collection, user1, user2 UserChoices) ([]Game, error) {
	var games []Game

	// Combine genres from both users
	combinedGenres := append(user1.Genres, user2.Genres...)

	// Query for games with genres matching those in combinedGenres
	cursor, err := collection.Find(context.Background(), bson.M{"genres.name": bson.M{"$in": combinedGenres}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var game Game
		err := cursor.Decode(&game)
		if err != nil {
			// Handle the error and continue to the next document
			log.Println("Error decoding document:", err)
			continue
		}
		games = append(games, game)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return games, nil
}


func GamesForOneUser(collection *mongo.Collection, user UserChoices) ([]Game, error) {
	var games []Game

	// Query for games with genres matching those in combinedGenres
	cursor, err := collection.Find(context.Background(), bson.M{"genres.name": bson.M{"$in": user.Genres}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var game Game
		err := cursor.Decode(&game)
		if err != nil {
			// Handle the error and continue to the next document
			log.Println("Error decoding document:", err)
			continue
		}
		games = append(games, game)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return games, nil
}

func GetSimilarGames(game Game, baseURL string) ([]string, error) {
	var nodes []*cdp.Node

	// Construir a URL de pesquisa com base no nome do jogo
	searchURL := fmt.Sprintf("%s?s=%s", baseURL, game.Name)

	// Configurar opções do Brave
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.ExecPath("/usr/bin/brave-browser"),
	)

	// Criar o contexto do Brave e o contexto de cancelamento
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Criar um novo contexto do Brave
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Navegar diretamente para a URL de pesquisa
	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(2*time.Second), // Aguardar 2 segundos para garantir que a página seja carregada

		// Exemplo de aguardar a resposta (pode ser apropriado para o seu caso)
		chromedp.Sleep(2*time.Second),

		// le os títulos de cada jogo .gp-loop-title
		chromedp.Nodes(".gp-loop-title a", &nodes, chromedp.ByQueryAll),
		//printa os títulos


		//fecha o navegador

	)


	if err != nil {
		log.Fatal(err)
		return nil, err
	}


	
// Extract names and limit to a maximum of 3
var names []string
for _, node := range nodes {
	// Append only the names
	names = append(names, node.Children[0].NodeValue)

	// Break the loop if the maximum limit is reached
	if len(names) >= 3 {
		break
	}
}


//close the browser
err = chromedp.Cancel(ctx)
if err != nil {
	log.Fatal(err)
	return nil, err
}


return names, nil
}


func Top5GamesRated(games []Game) ([]Game, error) {
	sort.Slice(games, func(i, j int) bool {
		return games[i].Rating > games[j].Rating
	})
	return games[:5], nil

}
	


func GamesFromMongoDB(collection *mongo.Collection) ([]Game, error) {
	var games []Game
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var game Game
		err := cursor.Decode(&game)
		if err != nil {
			// Handle the error and continue to the next document
			log.Println("Error decoding document:", err)
			continue
		}
		games = append(games, game)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return games, nil
}


func FilterGamesByGenreList(collection *mongo.Collection, genreList []string) ([]Game, error) {
	var games []Game
	cursor, err := collection.Find(context.Background(), bson.M{"genres.name": bson.M{"$in": genreList}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var game Game
		err := cursor.Decode(&game)
		if err != nil {
			// Handle the error and continue to the next document
			log.Println("Error decoding document:", err)
			continue
		}
		games = append(games, game)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return games, nil
}





func getAllGames() (*GameListResponse, error) {
	var allGames []Game
	url := fmt.Sprintf("https://api.rawg.io/api/games?key=%s", apiKey)

	for i := 0; i < 3; i++ {
		response, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Erro ao obter a lista de jogos. Código de status: %d", response.StatusCode)
		}

		var gameList GameListResponse
		err = json.NewDecoder(response.Body).Decode(&gameList)
		if err != nil {
			return nil, err
		}

		allGames = append(allGames, gameList.Results...)

		// Se não houver mais páginas, saia do loop
		if gameList.Next == "" {
			break
		}

		// Atualize a URL para a próxima página
		url = gameList.Next
	}

	return &GameListResponse{Results: allGames}, nil
}


func SaveUserPreferences(collection *mongo.Collection, username string, genres []string) error {
	// Crie um novo documento
	userChoices := UserChoices{
		Username: username,
		Genres:   genres,
	}

	// Insira o documento no banco de dados
	_, err := collection.InsertOne(context.Background(), userChoices)
	if err != nil {
		return err
	}

	return nil
}

