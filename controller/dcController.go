package controller

import (
	"fmt"
	"gameMatcher/service"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
    "github.com/joho/godotenv"
)

var (
    collection       *mongo.Collection
  
)


func StartDiscord(collection *mongo.Collection) {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Erro ao carregar as variáveis de ambiente:", err)
    }

    // Obter o token Discord a partir da variável de ambiente
    discordToken := os.Getenv("DISCORD_TOKEN")
    fmt.Println("Discord Token:", discordToken)

    sess, err := discordgo.New("Bot " + discordToken)
    if err != nil {
        log.Fatal("Erro ao criar a sessão Discord:", err)
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

    // Verifique se a mensagem começa com o prefixo desejado (por exemplo, "!comando")
    if strings.HasPrefix(m.Content, "!comando") {
        // Extrai os argumentos da mensagem
        args := strings.Fields(m.Content[len("!comando"):])

        // Verifique se há argumentos suficientes
        if len(args) < 1 {
            // Se não houver argumentos suficientes, envie uma mensagem de erro
            _, err := s.ChannelMessageSend(m.ChannelID, "Comando inválido. Utilize `!comando <argumentos>`.")
            if err != nil {
                log.Println("Erro ao enviar a mensagem de erro:", err)
            }
            return
        }

        // Analise o comando e execute a lógica correspondente
        switch args[0] {
        case "escolherpreferencias":
            // Verifique se há argumentos suficientes para o comando escolherpreferencias
            if len(args) < 2 {
                _, err := s.ChannelMessageSend(m.ChannelID, "Comando escolherpreferencias requer argumentos. Utilize `!comando escolherpreferencias <gêneros>`.")
                if err != nil {
                    log.Println("Erro ao enviar a mensagem de erro:", err)
                }
                return
            }

            // Extrai as preferências da mensagem
            preferencesInput := strings.Join(args[1:], " ")
            genres := parseGenres(preferencesInput)

            // Salva as preferências do usuário no banco de dados
            err := service.SaveUserPreferences(collection, m.Author.ID, genres)
            if err != nil {
                log.Println("Error saving user preferences:", err)
                return
            }

            GamesByGenre, err := service.FilterGamesByGenreList(collection, genres)
            if err != nil {
                log.Println("Error getting games by genres:", err)
            }

            if len(GamesByGenre) == 0 {
                log.Println("No games found for the user's genres.")
                return
            }

            topGames,err := service.Top5GamesRated(GamesByGenre)
         
            sendGenderGamesMessages(s, m.ChannelID, topGames)


           


            var similarGames []string
            for _, game := range topGames {
                similar, err := SimilarGames(game)
                if err != nil {
                    log.Println("Error getting similar games:", err)
                    continue
                }
                similarGames = append(similarGames, similar...)
            }
    
    
            if len(similarGames) == 0 {
                log.Println("No similar games found.")
                return
            }
    
                    // Se precisar, verifique o conteúdo de 'similarGames' aqui
                fmt.Println("Similar Games:", similarGames)
    
                // Agora envie as mensagens
                sendSimilarGamesMessages(s, m.ChannelID, similarGames)

        default:
            // Se o comando não for reconhecido, envie uma mensagem de erro
            _, err := s.ChannelMessageSend(m.ChannelID, "Comando desconhecido.")
            if err != nil {
                log.Println("Erro ao enviar a mensagem de erro:", err)
            }
        }

        return
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
    if r.Emoji.Name == "❤️" {
        userID := r.UserID

        genres, err := service.GetUserGenres(collection, userID)
        if err != nil {
            log.Println("Error getting user genres:", err)
            return
        }

        if len(genres) == 0 {
            log.Println("User has no genres.")
            return
        }

        games, err := service.FilterGamesByGenreList(collection, genres)
        if err != nil {
            log.Println("Error getting games by genres:", err)
            return
        }

        if len(games) == 0 {
            log.Println("No games found for the user's genres.")
            return
        }

        sendGenderGamesMessages(s, r.ChannelID, games)

     
        
    }
}


func sendGenderGamesMessages(s *discordgo.Session, channelID string, games []service.Game) {
    for _, game := range games {
        // Assuming there's a field named "Name" in the Game struct
        gameInfo := fmt.Sprintf("%s (Rating: %.2f)", game.Name, game.Rating)

        message := fmt.Sprintf("Jogo Do Mesmo Genêro: %s", gameInfo)
        _, err := s.ChannelMessageSend(channelID, message)
        if err != nil {
            log.Println("Erro ao enviar a mensagem:", err)
        }
    }
}



func sendSimilarGamesMessages(s *discordgo.Session, channelID string, games []string) {

    for _, game := range games {
        message := fmt.Sprintf("Jogo Similar: %s", game)
        _, err := s.ChannelMessageSend(channelID, message)
        if err != nil {
            log.Println("Erro ao enviar a mensagem:", err)
        }
    }
}

func SimilarGames(game service.Game) ([]string, error) {
    baseURL := "https://gameslikefinder.com/" // Substitua pelo seu URL base real
    similar, err := service.GetSimilarGames(game, baseURL)

    if err != nil {
        log.Fatal(err)
        // Lide com o erro adequadamente, por exemplo, retorne ou registre-o
        return nil, err
    }

    // Use a fatia 'similar' conforme necessário
    fmt.Println(similar)

    // Retorna os três primeiros jogos similares, se houver
    if len(similar) >= 3 {
        return similar[:3], nil
    }

    // Se houver menos de três jogos similares, retorne todos
    return similar, nil
}




// func GetSimilarGames(game Game, baseURL string) ([]string, error) {
// 	var nodes []*cdp.Node

// 	// Construir a URL de pesquisa com base no nome do jogo
// 	searchURL := fmt.Sprintf("%s?s=%s", baseURL, game.Name)

// 	// Configurar opções do Brave
// 	opts := append(chromedp.DefaultExecAllocatorOptions[:],
// 		chromedp.Flag("headless", false),
// 		chromedp.Flag("disable-gpu", true),
// 		chromedp.ExecPath("/usr/bin/brave-browser"),
// 	)

// 	// Criar o contexto do Brave e o contexto de cancelamento
// 	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
// 	defer cancel()

// 	// Criar um novo contexto do Brave
// 	ctx, cancel := chromedp.NewContext(allocCtx)
// 	defer cancel()

// 	// Navegar diretamente para a URL de pesquisa
// 	err := chromedp.Run(ctx,
// 		chromedp.Navigate(searchURL),
// 		chromedp.Sleep(2*time.Second), // Aguardar 2 segundos para garantir que a página seja carregada

// 		// Exemplo de aguardar a resposta (pode ser apropriado para o seu caso)
// 		chromedp.Sleep(2*time.Second),

// 		// le os títulos de cada jogo .gp-loop-title
// 		chromedp.Nodes(".gp-loop-title a", &nodes, chromedp.ByQueryAll),
// 		//printa os títulos


// 		//fecha o navegador

// 	)


// 	if err != nil {
// 		log.Fatal(err)
// 		return nil, err
// 	}


	
// // Extract names and limit to a maximum of 3
// var names []string
// for _, node := range nodes {
// 	// Append only the names
// 	names = append(names, node.Children[0].NodeValue)

// 	// Break the loop if the maximum limit is reached
// 	if len(names) >= 3 {
// 		break
// 	}
// }


// //close the browser
// err = chromedp.Cancel(ctx)
// if err != nil {
// 	log.Fatal(err)
// 	return nil, err
// }


// return names, nil
// }
