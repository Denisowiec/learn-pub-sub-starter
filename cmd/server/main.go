package main

import (
	"fmt"
	"log"

	"github.com/Denisowiec/learn-pub-sub-starter/internal/gamelogic"
	"github.com/Denisowiec/learn-pub-sub-starter/internal/pubsub"
	"github.com/Denisowiec/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.Println("Starting Peril server...")
	connAdr := "amqp://guest:guest@localhost:5672/"

	rabbitConn, err := amqp.Dial(connAdr)
	if err != nil {
		log.Fatal("Error connecting to the RabbitMQ server: ", err)
	}

	defer rabbitConn.Close()
	log.Println("Connection to RabbitMQ was successful.")

	// Creating a channel for pause/unpause messages
	pauseChan, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal("Error creating a AMQP channel: ", err)
	}

	_, _, err = pubsub.DeclareAndBind(
		rabbitConn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		fmt.Sprintf("%s.*", routing.GameLogSlug),
		pubsub.Durable)
	if err != nil {
		log.Fatal("Couldn't establish a queue for the game logs")
	}

	gamelogic.PrintServerHelp()
	// REPL loop
	repl := true
	for repl {
		input := gamelogic.GetInput()

		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "pause":
			log.Println("Sending a pause message to RabbitMQ...")

			err = pubsub.PublishJSON(pauseChan, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Println("Failed to publish on the pauseChan: ", err)
				continue
			}
			continue
		case "resume":
			log.Println("Sending a resume message to RabbitMQ...")

			err = pubsub.PublishJSON(pauseChan, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Println("Failed to publish on the pauseChan: ", err)
				continue
			}
			continue
		case "quit":
			log.Println("Server exiting...")
			repl = false
		case "":
			continue
		default:
			continue
		}
		if input[0] == "quit" {
		}
	}

}
