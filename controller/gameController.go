package controller

import (
	"encoding/json"
	"fmt"
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

	r.Run()
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
