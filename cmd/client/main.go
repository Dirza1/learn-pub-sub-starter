package main

import (
	"fmt"
	"os"
	"os/signal"

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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	sig := <-signalChan
	fmt.Printf("\nReceived %v, shutting down...\n", sig)

}
