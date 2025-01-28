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

	// Create a connection to RabbitMQ server
	conn, err := amqp.Dial(rabbitMQConnectionString)
	if err != nil {
		log.Fatalf("could not connect to the RabbitMQ server: %v\n", err)
	}
	defer conn.Close()
	log.Println("Peril game server connected to RabbitMQ!")

	// Create a channel to the RabbitMQ server
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v\n", err)
	}

	// Subscribe to the game log queue
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
		handlerLogs(),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue declared and bound!\n")

	// Print Help message
	gamelogic.PrintServerHelp()

	// Start REPL
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		if input[0] == "pause" {
			log.Println("Sending pause message")
			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
			log.Println("Pause message sent!")
		} else if input[0] == "resume" {
			log.Println("Sending resume message")
			channel, err := conn.Channel()
			if err != nil {
				log.Fatalf("could not create channel: %v\n", err)
			}

			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
			log.Println("Resume message sent!")
		} else if input[0] == "quit" {
			log.Println("Bye Bye!!")
			break
		} else {
			log.Println("Unknown Command")
			break
		}
	}
}
