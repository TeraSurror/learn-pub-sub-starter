package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rabbitMQConnectionString string = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitMQConnectionString)
	if err != nil {
		log.Fatalf("could not connect to the RabbitMQ server: %v\n", err)
	}
	defer conn.Close()
	log.Println("Peril game client connected to RabbitMQ!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Pause Queue %v declared and bound!\n", queue.Name)

	gameState := gamelogic.NewGameState(username)

	for {
		input := gamelogic.GetInput()

		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			err = gameState.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			_, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Move command successful")
		case "help":
			gamelogic.PrintClientHelp()
		case "status":
			gameState.CommandStatus()
		case "spam":
			fmt.Println("Spamming not supported yet")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("unknown command")
		}

	}
}
