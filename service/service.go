package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	"fmt"
	"log"
	_ "github.com/chromedp/cdproto/cdp"
	"github.com/gocolly/colly/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Game struct {
	Name   string  `json:"name"`
	Genres []Genre `json:"genres"`
	Rating float64 `json:"rating"`
}

type Genre struct {
	Name string `json:"name"`
}

type GameListResponse struct {
	Results []Game `json:"results"`
	Next    string `json:"next"`
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
	var names []string

	log.Print("Buscando jogos similares para", game.Name)

	// Criar um novo coletor
	c := colly.NewCollector()

	// Configurar o coletor
	c.OnHTML(".gp-loop-title a", func(e *colly.HTMLElement) {
		// Callback para tratar os elementos HTML correspondentes
		name := e.Text
		names = append(names, name)

		// Limite para um máximo de 3 jogos
		if len(names) >= 3 {
			c.Visit("") // Encerrar a coleta
		}
	})

	// Visitar a URL de pesquisa
	err := c.Visit(fmt.Sprintf("%s?s=%s", baseURL, game.Name))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Check if no similar games are found
	if len(names) == 0 {
		log.Println("No similar games found for", game.Name)
		return nil, nil
	}

	return names, nil
}

func Top5GamesRated(games []Game) ([]Game, error) {
    if len(games) < 5 {
        // Handle the case where there are less than 5 games
        return nil, errors.New("not enough games to get top 5")
    }

    // Sort games by rating in descending order
    sort.Slice(games, func(i, j int) bool {
        return games[i].Rating > games[j].Rating
    })

    // Return the top 5 games
    return games[:5], nil
}

//formatando os jogos do banco


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

func GetAllGames() (*GameListResponse, error) {
	var allGames []Game
	url := fmt.Sprintf("https://api.rawg.io/api/games?key=%s", apiKey)

	for i := 0; i < 100; i++ {
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

func GetUserGenres(collection *mongo.Collection, username string) ([]string, error) {
	// Query for the user
	var user UserChoices
	err := collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return user.Genres, nil
}



func FormatGameList(genres []string, games []string) []string {
	// Formatando os gêneros em uma string
	genresStr := strings.Join(genres, ", ")

	// Formatando os jogos em uma lista com uma linha de espaço para cada jogo
	var gamesList []string
	for _, game := range games {
		gamesList = append(gamesList, fmt.Sprintf("Jogos encontrados para os gêneros [%s]:\n%s", genresStr, game))
	}


	return gamesList
}

func FormatDBGamesList(games []Game) []string {
    var gamesList []string
    for i, game := range games {
        formattedGenres := formatGenres(game.Genres)
        gameEntry := fmt.Sprintf("%d. *%s\n   - Genres: %s\n   - Rating: %.2f", i+1, game.Name, formattedGenres, game.Rating)
        gamesList = append(gamesList, gameEntry)
    }
    return gamesList
}

func formatGenres(genres []Genre) string {
    var formattedGenres []string
    for _, genre := range genres {
        formattedGenres = append(formattedGenres, fmt.Sprintf("`%s`", genre.Name))
    }
    return strings.Join(formattedGenres, " ")
}