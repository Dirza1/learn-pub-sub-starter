package main

import (
	"fmt"
	"strings"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer connection.Close()
	fmt.Println("Connection successfull")
	pauseChan, err := connection.Channel()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer pauseChan.Close()

	logCha, logQue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.Durable)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer logCha.Close()
	fmt.Printf("declaured que: %s\n", logQue.Name)
	gamelogic.PrintServerHelp()

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch strings.ToLower(input[0]) {
		case "pause":
			fmt.Println("Pausing the server")
			if err := pubsub.PublishJSON(pauseChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
		case "resume":
			fmt.Println("Resuming the server")
			if err := pubsub.PublishJSON(pauseChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false}); err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
		case "quit":
			fmt.Println("Exiting the game")
			return
		default:
			fmt.Printf("Unknown command: %q\n", input[0])
		}
	}
}
