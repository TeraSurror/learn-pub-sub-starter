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
	log.Println("Peril game server connected to RabbitMQ!")

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v\n", err)
	}

	gamelogic.PrintServerHelp()

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
