package controller

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    // "strings"
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/joho/godotenv"
    _"gameMatcher/service"
)




func StartTelegram(collection *mongo.Collection) {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading environment variables:", err)
    }

    // Get the Telegram bot token from the environment variable
    telegramToken := os.Getenv("TELEGRAM_TOKEN")
    fmt.Println("Telegram Token:", telegramToken)

    bot, err := tgbotapi.NewBotAPI(telegramToken)
    if err != nil {
        log.Fatal("Error creating Telegram bot:", err)
    }

    // Set up your handlers, similar to Discord
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.GetUpdatesChan(u)

    // Handle updates in a separate goroutine
    go func() {
        for update := range updates {
            if update.Message != nil {
                // Handle incoming messages, similar to the Discord handler
              
            }
        }
    }()

    // Handle signals to gracefully shutdown the bot
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc
}


// func messageCreate(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, collection *mongo.Collection) {
//     // Similar logic as in the Discord version, but adapt for Telegram API
//     // You may need to handle different types of messages and commands in Telegram
// }
