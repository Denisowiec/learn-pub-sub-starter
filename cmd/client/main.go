package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Denisowiec/learn-pub-sub-starter/internal/gamelogic"
	"github.com/Denisowiec/learn-pub-sub-starter/internal/pubsub"
	"github.com/Denisowiec/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	// Establish connection to the RabbitMQ server
	connAdr := "amqp://guest:guest@localhost:5672/"

	rabbitConn, err := amqp.Dial(connAdr)
	if err != nil {
		log.Fatal("Error connecting to the RabbitMQ server: ", err)
	}

	defer rabbitConn.Close()
	log.Println("Connection to RabbitMQ was successful.")

	// Here starts actual client logic
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = pubsub.DeclareAndBind(
		rabbitConn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient)

	if err != nil {
		log.Fatal("Unable to declare a rabbitMQ queue: ", err)
	}

	state := gamelogic.NewGameState(username)

	repl := true
	for repl {
		input := gamelogic.GetInput()

		switch input[0] {
		case "spawn":
			err = state.CommandSpawn(input)
			if err != nil {
				log.Println(err)
				continue
			}
		case "move":
			_, err = state.CommandMove(input)
			if err != nil {
				log.Println(err)
				continue
			}
			fmt.Println("Move successful")
		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not alowed yet!")
		case "":
			continue
		default:
			fmt.Println("Command unknown")
			continue
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("Recieved interrupt signal. Exiting...")
}
