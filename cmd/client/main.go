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
	fmt.Println("Starting Peril client...")

	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer connection.Close()

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		return
	}
	queNname := routing.PauseKey + "." + userName
	PlayerChan, PlayerQue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, queNname, routing.PauseKey, pubsub.Transient)
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		return
	}
	defer PlayerChan.Close()
	fmt.Println("Declared queue:", PlayerQue.Name)

	gameState := gamelogic.NewGameState(userName)
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch strings.ToLower(input[0]) {
		case "spawn":
			gameState.CommandSpawn(input)
		case "move":
			gameState.CommandMove(input)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming is not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("Unknown command: %s", input[0])
		}
	}
}
