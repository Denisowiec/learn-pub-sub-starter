package main

import (
	"log"
	"os"
	"os/signal"

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

	err = pubsub.PublishJSON(pauseChan, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: true})
	if err != nil {
		log.Println("Failed to publish on the pauseChan: ", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("Recieved interrupt signal. Exiting...")
}
