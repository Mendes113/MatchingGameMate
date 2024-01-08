package controller

import (
	"fmt"
	"gameMatcher/service"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"

)

var userGenreChoices = make(map[int64][]string)

const (
    startCommand         = "start"
    findGamesByGenreCmd = "find_games_by_genre"
    findGamesCmd         = "find_games"
    telegramTokenEnv     = "TELEGRAM_TOKEN"
)


func StartTelegram(collection *mongo.Collection) {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading environment variables:", err)
    }

    telegramToken := os.Getenv(telegramTokenEnv)
    fmt.Println("Telegram Token:", telegramToken)

    bot, err := tgbotapi.NewBotAPI(telegramToken)
    if err != nil {
        log.Fatal("Error creating Telegram bot:", err)
    }

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.GetUpdatesChan(u)

   

    go func() {
        for update := range updates {
            if update.Message != nil {
                if update.Message.IsCommand() && update.Message.Command() == startCommand {
                    sendWelcomeMessage(bot, update.Message.Chat.ID)
                }
            } else if update.CallbackQuery != nil {
                handleCallbackQuery(update.CallbackQuery, collection, bot, userGenreChoices)
            }
        }
    }()

    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc
}

func sendWelcomeMessage(bot *tgbotapi.BotAPI, chatID int64) {
    welcomeMsg := tgbotapi.NewMessage(chatID, "Olá! Eu sou o GameMatcherBot🖥️! Vou te ajudar a encontrar jogos que você vai gostar 😊!")
    bot.Send(welcomeMsg)

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("Vamos Começar", findGamesByGenreCmd),
          
        ),
    )

    msg := tgbotapi.NewMessage(chatID, "Aqui vão algumas opções de gêneros de jogos que você pode escolher  🎮 :")
    msg.ReplyMarkup = keyboard

    bot.Send(msg)
}

var userSelectedGenres = make(map[int64][]string) // Map to store user selected genres

func handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, collection *mongo.Collection, bot *tgbotapi.BotAPI, userGenreChoices map[int64][]string) {
    var response string
	var selectedGenres []string
    switch callbackQuery.Data {
    case "start_conversation":
        // Process user genre choices
        genres := userGenreChoices[callbackQuery.Message.Chat.ID]
        response = fmt.Sprintf("Você escolheu os seguintes gêneros: %v", genres)
	case "find_games_by_genre":
        selectedGenres = optGamesByGenre(callbackQuery, bot, collection, userGenreChoices)
        response = fmt.Sprintf("Você escolheu os seguintes gêneros: %v", selectedGenres)
        userGenreChoices[callbackQuery.Message.Chat.ID] = selectedGenres  // Update userGenreChoices here
        log.Print("Selected genres: ", userGenreChoices[callbackQuery.Message.Chat.ID])

    case "filter_best_games":
        log.Println("Botão 'Buscar Melhores Jogos Do Genero' pressionado")
        log.Print("Selected genres: ", userGenreChoices[callbackQuery.Message.Chat.ID])

        if len(userGenreChoices[callbackQuery.Message.Chat.ID]) == 0 {
            response = "Por favor, escolha pelo menos um gênero antes de buscar jogos."
        } else {
            bestGames, err := service.FilterGamesByGenreList(collection, userGenreChoices[callbackQuery.Message.Chat.ID])
            if err != nil {
                fmt.Println(err)
                response = "Erro ao buscar jogos"
            } else {
                top5, err := service.Top5GamesRated(bestGames)
                if err != nil {
                    fmt.Println(err)
                    response = "Erro ao buscar jogos"
                } else {
                    formattedGames := service.FormatDBGamesList( top5)
                    response = fmt.Sprintf("Os 5 melhores jogos para os gêneros escolhidos são:\n%s", strings.Join(formattedGames, "\n"))
                }
            }
        }
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, response)
		bot.Send(msg)
	
		

	case "Action", "Adventure", "Strategy", "RPG", "Sports", "Puzzle":
		optGamesByGenre(callbackQuery, bot,  collection, userGenreChoices)
    case "find_games":
        // Handle "Buscar jogos similares"
        response = "Buscando jogos..."
        baseURL := "https://gameslikefinder.com"
        genres := userGenreChoices[callbackQuery.Message.Chat.ID]
        // Converter os gêneros para maiúsculas
        for i, genre := range genres {
            genres[i] = strings.ToUpper(genre)
        }
        games, err := service.FilterGamesByGenreList(collection, genres)
        gamesGenre, _ := service.Top5GamesRated(games)

        if err != nil {
            response = "Erro ao buscar jogos"
        } else {
            // Inicializa a lista de jogos encontrados
            var foundGames []string

            // Para cada jogo no gênero, obtém jogos similares
            for _, game := range gamesGenre {
                similarGames, err := service.GetSimilarGames(game, baseURL)
                if err != nil {
                    response = fmt.Sprintf("Erro ao buscar jogos similares para %s", game.Name)
                } else {
                    foundGames = append(foundGames, similarGames...)
                }
            }

            foundGames = service.FormatGameList(genres, foundGames)
            response = fmt.Sprintf("Jogos encontrados para os gêneros %v: %v", genres, foundGames)
        }

   

        msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, response)
       
        bot.Send(msg)
}
}



func optGamesByGenre(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, collection *mongo.Collection, userGenreChoices map[int64][]string) []string {
	log.Printf("Before updating userGenreChoices: %v", userGenreChoices)

	// Handle "Buscar jogos por gênero"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Action", "Action"),
			tgbotapi.NewInlineKeyboardButtonData("Adventure", "Adventure"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Strategy", "Strategy"),
			tgbotapi.NewInlineKeyboardButtonData("RPG", "RPG"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Sports", "Sports"),
			tgbotapi.NewInlineKeyboardButtonData("Puzzle", "Puzzle"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Buscar Melhores Jogos Do Genero", "filter_best_games"),
		),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("Buscar Jogos Similares Para o Genero", "find_games"),
        ),
	)

	// switch case for each genre
	var genreMap = map[string]string{
		"Action":   "Action",
		"Adventure": "Adventure",
		"Strategy":  "Strategy",
		"RPG":       "RPG",
		"Sports":    "Sports",
		"Puzzle":    "Puzzle",
	}

	var selectedGenres []string

	if genre, ok := genreMap[callbackQuery.Data]; ok {
		// Convert to uppercase before adding to userGenreChoices
		uppercaseGenre := strings.ToUpper(genre)
		userGenreChoices[callbackQuery.Message.Chat.ID] = append(userGenreChoices[callbackQuery.Message.Chat.ID], uppercaseGenre)
		selectedGenres = append(selectedGenres, uppercaseGenre)
	}

	response := "Escolha um gênero para buscar jogos:"
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, response)
	msg.ReplyMarkup = keyboard

	log.Printf("After updating userGenreChoices: %v", userGenreChoices)

	// Sending the keyboard to the user
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Error sending message:", err)
	}

	return selectedGenres
}



