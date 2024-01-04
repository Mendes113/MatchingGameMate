package controller

import (
	"fmt"
	"gameMatcher/service"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
)

// ...
func startDiscord(collection *mongo.Collection) {
    sess, err := discordgo.New("Bot MTE5MjI3Njg4NjUyOTc3NzY5NA.G_NVmG.nTwn--HtXR1wwI0uWh6JmTO2X_cb3nHDjBlTBg")
    if err != nil {
        log.Fatal(err)
    }

    sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
        messageCreate(s, m, collection)
    })
    sess.AddHandler(messageReactionAdd)

    sess.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

    err = sess.Open()
    if err != nil {
        log.Fatal(err)
    }

    defer sess.Close()

    fmt.Println("Bot is running...")

    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc
}



func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate, collection *mongo.Collection) {
    // Verifique se o autor da mensagem é o bot para evitar loops infinitos
    if m.Author.ID == s.State.User.ID {
        return
    }

    // Verifique se a mensagem é uma solicitação para escolher as preferências de jogo
    if m.Content == "!escolherpreferencias" {
        // Envie uma mensagem inicial
        _, err := s.ChannelMessageSend(m.ChannelID, "Vamos escolher suas preferências de jogo! Por favor, responda com seus gêneros favoritos, separados por vírgulas.")
        if err != nil {
            log.Println("Error sending message:", err)
            return
        }

        // Aguarde a resposta do usuário usando um filtro de mensagem
        responseCh := make(chan *discordgo.MessageCreate, 1)

        // separa os gêneros fornecidos pelo usuário
        go func() {
            for {
                select {
                case msg := <-responseCh:
                    // Analise os gêneros fornecidos pelo usuário
                    genres := parseGenres(msg.Content)

                    // Salve as preferências do usuário no banco de dados
                    err := service.SaveUserPreferences(collection, msg.Author.ID, genres)
                    if err != nil {
                        log.Println("Error saving user preferences:", err)
                        return
                    }

                    // Envie uma mensagem de confirmação
                    _, err = s.ChannelMessageSend(msg.ChannelID, "Suas preferências foram salvas com sucesso!")
                    if err != nil {
                        log.Println("Error sending message:", err)
                        return
                    }

                    return
                case <-time.After(60 * time.Second):
                    _, err := s.ChannelMessageSend(m.ChannelID, "Tempo limite atingido. Por favor, tente novamente.")
                    if err != nil {
                        log.Println("Error sending message:", err)
                        return
                    }
                    return
                }
            }
        }()
    }
}

func parseGenres(input string) []string {
    // Separe os gêneros usando vírgulas
    genres := strings.Split(input, ",")

    // Remova espaços em branco desnecessários em cada gênero
    for i, genre := range genres {
        genres[i] = strings.TrimSpace(genre)
    }

    // Filtra elementos vazios
    genres = filterEmptyStrings(genres)

    return genres
}

// Função auxiliar para filtrar elementos vazios de uma lista de strings
func filterEmptyStrings(slice []string) []string {
    var result []string
    for _, s := range slice {
        if s != "" {
            result = append(result, s)
        }
    }
    return result
}

func messageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
    // Your existing reaction add logic here
}