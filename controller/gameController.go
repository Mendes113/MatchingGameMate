package controller

import (
	"encoding/json"
	"fmt"
	"gameMatcher/model"
	"gameMatcher/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// StartServer starts the gin server
func StartServer(collection *mongo.Collection) {
	
	r := gin.Default()
	r.GET("/games", func(c *gin.Context) {
		getGames(c, collection)
	})

	r.POST("/games", func(c *gin.Context) {
		GetSimilarGamesTwoUsers(c, collection)
	})

	r.GET("/all", func(c *gin.Context) {
		getGamesFromAPI(c)
	})

	r.GET("getid", func(c *gin.Context){
		GetSteamGames(c)
	})

	r.Run(":8081")
}

func getGames(c *gin.Context, collection *mongo.Collection) {
	games, err := service.GamesFromMongoDB(collection)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(200, games)
}

func GetSimilarGamesTwoUsers(c *gin.Context, collection *mongo.Collection) {
    var data map[string]interface{}

    if err := c.ShouldBindJSON(&data); err != nil {
        fmt.Println(err)
        c.JSON(400, gin.H{"error": "Bad Request"})
        return
    }

    // Converte o mapa para os structs user1 e user2
    user1JSON, _ := json.Marshal(data["user1"])
    user2JSON, _ := json.Marshal(data["user2"])

    var user1 service.UserChoices
    var user2 service.UserChoices

    json.Unmarshal(user1JSON, &user1)
    json.Unmarshal(user2JSON, &user2)

    games, err := service.EqualGamesIn2Users(collection, user1, user2)
    if err != nil {
        fmt.Println(err)
        c.JSON(500, gin.H{"error": "Internal Server Error"})
        return
    }

	top5,err := service.Top5GamesRated(games)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}



    c.JSON(200, top5)
}


func getGamesFromAPI(c *gin.Context) {
	gamesResponse, err := service.GetAllGames()
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}

	// Extract the relevant information from gamesResponse and create a []model.Game
	var games []model.Game
for _, gameItem := range gamesResponse.Results {
	game := model.Game{
		Name:   gameItem.Name,
		Rating: gameItem.Rating,
	}

	// Convert service.Genre to model.Genre
	for _, genreItem := range gameItem.Genres {
		game.Genres = append(game.Genres, ConvertServiceGenreToModelGenre(genreItem))
	}

	games = append(games, game)
}

	// Save the extracted games to MongoDB
	model.SaveGamesInMongoDB(games, model.GetCollection("games"))
	fmt.Println("Games saved to MongoDB!")

	c.JSON(200, games)
}

type Game struct {
	Name    string   `json:"name"`
	Genres  []Genre  `json:"genres"`
	Rating  float64  `json:"rating"`
  }
  
  type Genre struct {
	Name string `json:"name"`
  }

  func ConvertServiceGenreToModelGenre(serviceGenre service.Genre) model.Genre {
	return model.Genre{
		Name: serviceGenre.Name,
	}
}



func GetSteamGames(c *gin.Context) {
	// Extrair o nome do corpo da solicitação JSON

	service.GetGamesFromSteamAPI1(model.GetCollection("steamGames"))
}