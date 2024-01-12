package model

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)




var (
	mongoClient *mongo.Client
	once        sync.Once
)

// getMongoClient returns a MongoDB client using a singleton pattern.
func getMongoClient() *mongo.Client {
	once.Do(func() {
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
		mongoClient = client
	})

	return mongoClient
}

// MongoConnection returns the MongoDB client using a singleton pattern.
func MongoConnection() *mongo.Client {
	return getMongoClient()
}

func GetCollection(collectionName string) *mongo.Collection {
	mongoClient := MongoConnection()
	collection := mongoClient.Database("gamestore").Collection(collectionName)
	if collection == nil {
		log.Fatal("Collection is nil")
	}
	return collection
}


func GetCollectionFromDB(dbName, collectionName string) *mongo.Collection {
	mongoClient := MongoConnection()
	collection := mongoClient.Database(dbName).Collection(collectionName)
	if collection == nil {
		log.Fatal("Collection is nil")
	}
	return collection
}

func SaveGamesInMongoDB(games []Game, collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var gamesInterface []interface{}

	for _, game := range games {
		for i := range game.Genres {
			game.Genres[i].Name = strings.ToUpper(game.Genres[i].Name)
		}
		gamesInterface = append(gamesInterface, game)
	}

	for _, game := range gamesInterface {
		filter := bson.M{"name": game.(Game).Name}
		var result Game
		err := collection.FindOne(ctx, filter).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				continue
			}
			return fmt.Errorf("error checking if game already exists: %v", err)
		}
		return fmt.Errorf("game '%s' already in the database", result.Name)
	}

	_, err := collection.InsertMany(ctx, gamesInterface)
	if err != nil {
		return fmt.Errorf("error inserting games into the database: %v", err)
	}

	return nil
}


type Game struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name   string             `json:"name"`
	Genres []Genre            `json:"genres"`
	Rating float64            `json:"rating"`
}

  type Genre struct {
	Name string `json:"name"`
  }

type GameListResponse struct {
	Results []Game `json:"results"`
	Next    string  `json:"next"`
}

func GetGenresFromDB(collection *mongo.Collection) ([]string, error) {
    var genres []string

    cursor, err := collection.Find(context.Background(), bson.D{})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())

    for cursor.Next(context.Background()) {
        var document bson.M
        if err := cursor.Decode(&document); err != nil {
            log.Println("Error decoding document:", err)
            continue
        }

        // Assume that the field containing genre information is "genres" and it's an array of objects
        if genreArray, ok := document["genres"].([]interface{}); ok {
            for _, genreObject := range genreArray {
                if genreMap, ok := genreObject.(map[string]interface{}); ok {
                    if genreName, exists := genreMap["name"].(string); exists {
                        // Make an exception for "RPG" to be all uppercase
                        if strings.ToUpper(genreName) == "RPG" {
                            genres = append(genres, "RPG")
                        } else {
                            // Capitalize the first letter of other genres
                            capitalizedGenre := strings.Title(genreName)
                            genres = append(genres, capitalizedGenre)
                        }
                    }
                }
            }
        }
    }

    if err := cursor.Err(); err != nil {
        return nil, err
    }

    return genres, nil
}


func containsGenre(genres []string, genre string) bool {
    for _, g := range genres {
        if g == genre {
            return true
        }
    }
    return false
}


func UpdateGenresToUppercase(collection *mongo.Collection) error {
    cursor, err := collection.Find(context.Background(), bson.D{})
    if err != nil {
        return err
    }
    defer cursor.Close(context.Background())

    for cursor.Next(context.Background()) {
        var game Game
        if err := cursor.Decode(&game); err != nil {
            log.Println("Error decoding document:", err)
            continue
        }

        // Atualizar os nomes dos gêneros para maiúsculas
        for i := range game.Genres {
            game.Genres[i].Name = strings.ToUpper(game.Genres[i].Name)
        }

        // Atualizar o documento no MongoDB
        filter := bson.M{"_id": game.ID}
        update := bson.M{"$set": bson.M{"genres": game.Genres}}
        _, err := collection.UpdateOne(context.Background(), filter, update)
        if err != nil {
            log.Println("Error updating document:", err)
        }
    }

    if err := cursor.Err(); err != nil {
        return err
    }

    return nil
}


type SteamGame struct {
	AppID int `json:"appid"`
	Name string   `json:"name"`
}

func SaveSteamResponse(collection *mongo.Collection, games []SteamGame) error {
	// Crie um contexto
	ctx := context.TODO()

	// Converte a lista de jogos SteamGame para um array de documentos BSON
	var documents []interface{}
	for _, game := range games {
		document := bson.D{
			{Key: "appID", Value: game.AppID},
			{Key: "name", Value: game.Name},
		}
		documents = append(documents, document)
	}

	// Insere os documentos na coleção
	_, err := collection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("erro ao inserir na coleção: %v", err)
	}

	return nil
}



func GetSteamGameIdUsingName(collection *mongo.Collection, ctx context.Context, name string) (SteamGame, error) {
    filter := bson.M{"name": name}
    var result SteamGame

    err := collection.FindOne(ctx, filter).Decode(&result)
    if err != nil {
        notFoundError := fmt.Errorf("jogo com o nome '%s' não encontrado", name)
        genericError := fmt.Errorf("erro ao buscar jogo: %v", err)

        if err == mongo.ErrNoDocuments {
            log.Print("Jogo não encontrado")
            return SteamGame{}, notFoundError
        }

        log.Printf("Erro ao buscar jogo: %v", err)
        return SteamGame{}, genericError
    }

    return result, nil
}
