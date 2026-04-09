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
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()
	newChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	err = pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.SimpleQueueDurable, handlerLogs())
	if err != nil {
		log.Fatalf("Failed to subscribe to game logs: %s", err)
	}

	fmt.Println("Peril game server connected to RabbitMQ!")
	gamelogic.PrintServerHelp()
	for {
		userWords := gamelogic.GetInput()
		if len(userWords) == 0 {
			continue
		}
		first := userWords[0]
		switch first {
		case "pause":
			log.Println("Pausing Peril game server")
			err = pubsub.PublishJSON(newChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
		case "resume":
			log.Println("Resuming Peril game server")
			err = pubsub.PublishJSON(newChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
		case "quit":
			log.Println("Quitting Peril game server")
			return
		default:
			log.Println("Unknown command")
		}
	}
}
