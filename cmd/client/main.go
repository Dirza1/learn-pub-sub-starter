package main

import (
	"fmt"
	"strconv"
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
	err = pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, queNname, routing.PauseKey, pubsub.Transient, handlerPause(gameState))
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		return
	}

	handleMove := func(m gamelogic.ArmyMove) {
		outcome := gameState.HandleMove(m)
		_ = outcome
		fmt.Print("> ")
	}

	err = pubsub.SubscribeJSON(connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+userName,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handleMove,
	)
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		return
	}
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
			id, _ := strconv.Atoi(input[2])
			currentUnit, _ := gameState.GetUnit(id)
			err = pubsub.PublishJSON(PlayerChan, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+userName, gamelogic.ArmyMove{
				Player:     gameState.Player,
				ToLocation: gamelogic.Location(input[1]),
				Units:      []gamelogic.Unit{currentUnit},
			})
			if err != nil {
				fmt.Printf("Err: %s\n", err)
				return
			}
			fmt.Println("Move published successfully")
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Println("> ")
		gs.HandlePause(ps)
	}
}
