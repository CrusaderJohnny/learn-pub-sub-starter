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
	fmt.Println("Starting Peril client...")
	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}
	gs := gamelogic.NewGameState(userName)
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, "pause."+userName, routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gs))
	if err != nil {
		log.Fatal(err)
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "army_moves."+userName, "army_moves.*", pubsub.SimpleQueueTransient, handlerMove(gs))
	if err != nil {
		log.Fatal(err)
	}
	moveCh, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	//REPL
	for {
		commandInput := gamelogic.GetInput()
		if len(commandInput) == 0 {
			continue
		}
		switch commandInput[0] {
		case "spawn":
			err = gs.CommandSpawn(commandInput)
			if err != nil {
				log.Println(err)
				continue
			}
		case "move":
			move, err := gs.CommandMove(commandInput)
			if err != nil {
				log.Println(err)
				continue
			}
			err = pubsub.PublishJSON(moveCh, routing.ExchangePerilTopic, "army_moves."+userName, move)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("Moved to", commandInput[1])
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			log.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Println("Unknown command")
			continue
		}
	}
}
