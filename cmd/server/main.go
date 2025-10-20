package main

import (
	"fmt"
	"os"
	"os/signal"

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
	pauzeChan, err := connection.Channel()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = pubsub.PublishJSON(pauzeChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	if err != nil {
		fmt.Printf("Err: %s/n", err)
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	sig := <-signalChan
	fmt.Printf("\nReceived %v, shutting down...\n", sig)
}
