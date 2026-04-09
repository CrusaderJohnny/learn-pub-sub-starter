package main

import (
	"fmt"
	"log"
	"strconv"

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
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, "pause."+userName, routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gs))
	if err != nil {
		log.Fatal(err)
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "army_moves."+userName, "army_moves.*", pubsub.SimpleQueueTransient, handlerMove(gs, publishCh))
	if err != nil {
		log.Fatal(err)
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "war", routing.WarRecognitionsPrefix+".*", pubsub.SimpleQueueDurable, handlerWar(gs, publishCh))

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
			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, "army_moves."+userName, move)
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
			if len(commandInput) < 2 {
				log.Println("spam requires at least two arguments")
				continue
			}
			spamAmount, err := strconv.Atoi(commandInput[1])
			if err != nil {
				log.Println(err)
				continue
			}
			for i := 0; i < spamAmount; i++ {
				err = pubsub.PublishGob(publishCh, routing.ExchangePerilTopic, routing.GameLogSlug+"."+userName, gamelogic.GetMaliciousLog())
				if err != nil {
					log.Println(err)
					continue
				}
			}
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Println("Unknown command")
			continue
		}
	}
}
