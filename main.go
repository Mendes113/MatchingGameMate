package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"


	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const apiKey = "3525e0e6ab9a480dbb58207514234680"

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


func main() {


		mongoClient := mongoConnection()
		defer mongoClient.Disconnect(context.Background())

		collection := mongoClient.Database("gamestore").Collection("games")

		// games, err := getAllGames()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// for _, game := range games.Results {
		// 	_, err := collection.InsertOne(context.Background(), game)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// }

		

		// for _, game := range gamesFromDB {
		// 	fmt.Println(game.Name)
		// }

		
		

		user1 := UserChoices{
			Username: "user1",
			Genres:   []string{"Action", "Adventure", "RPG"},}
		user2 := UserChoices{
			Username: "user2",
			Genres:   []string{"Action", "Adventure", "Strategy"},}
			
		equalGames, err := equalGamesIn2Users(collection, user1, user2)
		if err != nil {
			log.Fatal(err)
		}



		//best games for the 2 users
		top5Games, err := top5GamesRated(equalGames)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Top 5 games:")
		for _, game := range top5Games {
			fmt.Println(game.Name)
		}
		




} 



func top5GamesRated(games []Game) ([]Game, error) {
	sort.Slice(games, func(i, j int) bool {
		return games[i].Rating > games[j].Rating
	})
	return games[:5], nil

}
	


func gamesFromMongoDB(collection *mongo.Collection) ([]Game, error) {
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


func filterGamesByGenreList(collection *mongo.Collection, genreList []string) ([]Game, error) {
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


func mongoConnection() *mongo.Client {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	return client
}




func equalGamesIn2Users(collection *mongo.Collection, user1, user2 UserChoices) ([]Game, error) {
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