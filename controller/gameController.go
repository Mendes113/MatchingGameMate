package controller

import (
	"encoding/json"
	"fmt"
	"gameMatcher/model"
	"gameMatcher/service"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/mongo"
)



type ServerConfig struct {
	Collection *mongo.Collection
	Port       string
	// Add more configuration options as needed
}

func StartServer(config ServerConfig) {
	app := fiber.New()
	app.Use(logger.New())    // Middleware for logging
	// app.Use(recover.New())  // Middleware for error handling
	app.Post("/getgameid", GetGameIdByNameFiber)
	app.Get("/all", getGamesFromAPIFiber)
	// app.Get("/games", func(c *fiber.Ctx) error {
	// 	getGames(c, config.Collection)
	// 	return nil
	// })
	app.Post("/games", func(c *fiber.Ctx) error {
		GetSimilarGamesTwoUsers(c)
		return nil
	})

	app.Post("/getreviews", func(c *fiber.Ctx) error {
		GetReviews(c)
		return nil
	})


	err := app.Listen(config.Port)
	if err != nil {
		log.Fatal(err)
	}
}
	

// // StartServer starts the gin server
// func StartServer(collection *mongo.Collection) {
	
// 	r := gin.Default()
// 	r.GET("/games", func(c *gin.Context) {
// 		getGames(c, collection)
// 	})

// 	r.POST("/games", func(c *gin.Context) {
// 		GetSimilarGamesTwoUsers(c, collection)
// 	})

// 	r.GET("/all", func(c *gin.Context) {
// 		getGamesFromAPI(c)
// 	})

// 	r.GET("getid", func(c *gin.Context){
// 		GetSteamGames(c)
// 	})

// 	r.GET("getgameid", func(c *gin.Context){
// 		GetGameIdByName(c)
// 	})

// 	r.Run(":8081")
// }

func getGames(c *gin.Context, collection *mongo.Collection) {
	games, err := service.GamesFromMongoDB(collection)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(200, games)
}
func GetSimilarGamesTwoUsers(c *fiber.Ctx) error {
    var data map[string]interface{}

    if err := c.BodyParser(&data); err != nil {
        fmt.Println(err)
        return c.Status(400).JSON(fiber.Map{"error": "Bad Request"})
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
        return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
    }

    top5, err := service.Top5GamesRated(games)
    if err != nil {
        fmt.Println(err)
        return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
    }

    return c.JSON(top5)
}



func getGamesFromAPIFiber(c *fiber.Ctx) error {
	gamesResponse, err := service.GetAllGames()
	if err != nil {
		fmt.Println(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
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

	return c.JSON(games)
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



func GetSteamGames(c *fiber.Ctx) {
	// Extrair o nome do corpo da solicitação JSON

	service.GetGamesFromSteamAPI1(model.GetCollection("steamGames"))
}


// func GetGameIdByName(c *fiber.Ctx) {
//     // Defina a estrutura para o corpo da solicitação JSON
//     type RequestBody struct {
//         Name string `json:"name" binding:"required"`
//     }

//     // Extrair o nome do corpo da solicitação JSON
//     var requestBody RequestBody
//     if err := c.ShouldBindJSON(&requestBody); err != nil {
//         fmt.Println(err)
//         c.JSON(400, gin.H{"error": "Bad Request"})
//         return
//     }

//     // Agora você pode acessar o nome diretamente usando requestBody.Name
//     game := service.Game{Name: requestBody.Name}

//     gameId, err := service.GetSteamGameIdUsingName(model.GetCollection("steamGames"), game.Name)
//     if err != nil {
//         fmt.Println(err)
//         c.JSON(500, gin.H{"error": "Internal Server Error"})
//         return
//     }

//     c.JSON(200, gameId)
// }


type RequestBody struct {
    Name string `json:"name" binding:"required"`
}

func GetGameIdByNameFiber(c *fiber.Ctx) error {
    // Extrair o nome do corpo da solicitação JSON
    var requestBody RequestBody
    if err := c.BodyParser(&requestBody); err != nil {
        fmt.Println(err)
        return c.Status(400).JSON(fiber.Map{"error": "Bad Request"})
    }

    // Agora você pode acessar o nome diretamente usando requestBody.Name
    game := service.Game{Name: requestBody.Name}

    gameId, err := service.GetSteamGameIdUsingName(model.GetCollection("steamGames"), game.Name)
    if err != nil {
        fmt.Println(err)
        return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
    }

    return c.JSON(fiber.Map{"gameId": gameId})
}
func GetReviews(c *fiber.Ctx) error {
	// Extrair o nome do corpo da solicitação JSON
	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Bad Request"})
	}

	// Agora você pode acessar o nome diretamente usando requestBody.Name
	game := service.SteamGame{Name: requestBody.Name}

	// Converta service.Game para service.SteamGame diretamente
	steamGame := service.SteamGame(game)

	reviews, err := service.ScrapReviewsFromSteam(steamGame, model.GetCollection("steamGames"))
	if err != nil {
		fmt.Println(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error"})
	}

	return c.JSON(fiber.Map{"reviews": reviews})
}

