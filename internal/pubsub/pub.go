package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable   SimpleQueueType = iota
	Transient SimpleQueueType = iota
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	dat, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        dat,
	})
	if err != nil {
		return err
	}
	return nil
}

func createQueue(conn *amqp.Connection, queueName string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	// All durable queues will be non-auto-delete and non-exclusive. Transient will be the other way round
	dur := (queueType == Durable)
	autoDel := (queueType == Transient)
	excl := autoDel

	queue, err := channel.QueueDeclare(queueName, dur, autoDel, excl, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return channel, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T),
) error {
	channel, queue, err := createQueue(conn, queueName, queueType)
	if err != nil {
		return err
	}

	deliveries, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	go func() {
		for item := range deliveries {
			var content T
			err = json.Unmarshal(item.Body, &content)
			if err != nil {
				log.Fatal("Error unmarshalling server message")
			}

			handler(content)
			item.Ack(false)
		}
	}()

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, queue, err := createQueue(conn, queueName, queueType)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, nil
	}

	return channel, queue, nil
}
