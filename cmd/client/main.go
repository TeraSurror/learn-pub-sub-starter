package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rabbitMQConnectionString string = "amqp://guest:guest@localhost:5672/"

	// Create connection to RabbitMQ server
	conn, err := amqp.Dial(rabbitMQConnectionString)
	if err != nil {
		log.Fatalf("could not connect to the RabbitMQ server: %v\n", err)
	}
	defer conn.Close()
	log.Println("Peril game client connected to RabbitMQ!")

	// Create a channel to RabbitMQ server
	publishChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	// Print Welcome message
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	// Create new game state for player
	gameState := gamelogic.NewGameState(username)

	// Subscribe to Move queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, publishChannel),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}

	// Subscribe to Pause queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	// Subscribe to War queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gameState, publishChannel),
	)
	if err != nil {
		log.Fatalf("could not subscribe to war declarations: %v", err)
	}

	// Start REPL
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
			move, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				publishChannel,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+move.Player.Username,
				move,
			)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}
			fmt.Printf("Moved %v units to %s\n", len(move.Units), move.ToLocation)
		case "help":
			gamelogic.PrintClientHelp()
		case "status":
			gameState.CommandStatus()
		case "spam":
			if len(input) < 2 {
				fmt.Println("spam command must have word to work")
				continue
			}
			n, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Printf("error: %v\n", err)
				continue
			}
			for i := 0; i < n; i++ {
				malLog := gamelogic.GetMaliciousLog()
				err = publishGameLog(publishChannel, username, malLog)
				if err != nil {
					fmt.Printf("error: %v\n", err)
				}
			}

		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("unknown command")
		}

	}
}

func publishGameLog(publishChannel *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		publishChannel,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     msg,
		},
	)
}
