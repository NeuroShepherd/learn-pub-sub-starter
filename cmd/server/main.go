package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connAddr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connAddr)
	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %s\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to RabbitMQ")

	// new channel
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Failed to open a channel: %s\n", err)
		return
	}

	gamelogic.PrintServerHelp()

	for {
		word := gamelogic.GetInput() //returns slice of strings

		if len(word) == 0 {
			continue
		}

		switch word[0] {
		case "pause":
			fmt.Println("Pausing the game...")
			pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true, // false for resume
				},
			)
		case "resume":
			fmt.Println("Resuming the game...")
			pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false, // true for pause
				},
			)
		case "quit":
			fmt.Println("Quitting the server...")
			return
		case "help":
			gamelogic.PrintServerHelp()
		default:
			fmt.Printf("Unknown command: %s\n", word[0])
			gamelogic.PrintServerHelp()
		}
	}

}
