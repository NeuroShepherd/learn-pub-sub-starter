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
		fmt.Printf("Failed to connect client to RabbitMQ: %s\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Error during client welcome: %s\n", err)
		return
	}
	fmt.Printf("Welcome, %s!\n", username)

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON[routing.PlayingState](
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.QueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		fmt.Printf("Failed to subscribe to pause messages: %s\n", err)
		return
	}
	fmt.Println("Subscribed to pause messages")

	err = pubsub.SubscribeJSON[gamelogic.ArmyMove](
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, "*"),
		pubsub.QueueTransient,
		handlerMove(gameState),
	)
	if err != nil {
		fmt.Printf("Failed to subscribe to army moves messages: %s\n", err)
		return
	}
	fmt.Println("Subscribed to army_moves messages")

	// new channel
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Failed to open a channel: %s\n", err)
		return
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			fmt.Println(words[1:])
			err := gameState.CommandSpawn(
				words, // pass the rest of the words as arguments
			)
			if err != nil {
				fmt.Printf("Error spawning entity: %s\n", err)
				continue
			}
		case "move":
			armyMove, err := gameState.CommandMove(
				words, // pass the rest of the words as arguments
			)
			if err != nil {
				fmt.Printf("Error moving units: %s\n", err)
				continue
			}

			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
				armyMove,
			)
			if err != nil {
				fmt.Printf("Error publishing move: %s\n", err)
				continue
			}
			fmt.Print("Move command sent successfully.\n")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command. Type 'help' for a list of commands.")
			continue
		}
	}

}
