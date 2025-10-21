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

func declareBindSubscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType pubsub.SimpleQueueType,
	handler func(T),
) error {
	_, _, err := pubsub.DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)

	if err != nil {
		return err
	}

	err = pubsub.SubscribeJSON(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
	)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	fmt.Println("Starting Peril client...")
	// Establish connection to the RabbitMQ server
	connAdr := "amqp://guest:guest@localhost:5672/"

	rabbitConn, err := amqp.Dial(connAdr)
	if err != nil {
		log.Fatal("Error connecting to the RabbitMQ server: ", err)
	}

	defer rabbitConn.Close()

	channel, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal("Error establishing a RabbitMQ channel: ", err)
	}

	log.Println("Connection to RabbitMQ was successful.")

	// Here starts actual client logic
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	state := gamelogic.NewGameState(username)

	// Subscribing to the pause feature
	pauseQName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	err = declareBindSubscribe(rabbitConn, routing.ExchangePerilDirect, pauseQName, routing.PauseKey, pubsub.Transient, HandlerPause(state))
	if err != nil {
		log.Fatal("Unable to subscribe to a RabbitMQ queue: ", err)
	}
	// Subscribing to the army move queue
	armyMoveQName := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
	armyMoveKey := fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)

	err = declareBindSubscribe(rabbitConn, routing.ExchangePerilTopic, armyMoveQName, armyMoveKey, pubsub.Transient, HandlerMove(state))
	if err != nil {
		log.Fatal("Unable to subscribe to a RabbitMQ queue: ", err)
	}

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
			am, err := state.CommandMove(input)
			pubsub.PublishJSON(channel, routing.ExchangePerilTopic, armyMoveKey, am)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("Move published successfully")
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
