package main

import (
	"fmt"
	"time"

	"github.com/Denisowiec/learn-pub-sub-starter/internal/pubsub"
	"github.com/Denisowiec/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGameLog(ch *amqp.Channel, msg, username string) error {
	key := fmt.Sprintf("%s.%s", routing.GameLogSlug, username)
	formattedLog := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     msg,
		Username:    username,
	}
	err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, key, formattedLog)
	return err
}
