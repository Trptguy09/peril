package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	QueueDurable SimpleQueueType = iota
	QueueTransient
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("ERROR: Failed to open channel: %v", err) // Add logging for channel error
		return nil, amqp.Queue{}, err
	}

	durable := queueType == QueueDurable
	autoDelete := queueType == QueueTransient
	exclusive := queueType == QueueTransient

	log.Printf("DEBUG: Declaring queue '%s' with durable=%t, autoDelete=%t, exclusive=%t", queueName, durable, autoDelete, exclusive) // Keep this
	q, err := ch.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		false,
		nil,
	)
	if err != nil {
		log.Printf("ERROR: Failed to declare queue '%s': %v", queueName, err) // Add this logging
		return nil, amqp.Queue{}, err
	}
	log.Printf("DEBUG: Queue '%s' declared successfully.", q.Name) // Add this

	log.Printf("DEBUG: Binding queue '%s' to exchange '%s' with key '%s'", q.Name, exchange, key) // Add this
	err = ch.QueueBind(
		q.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		log.Printf("ERROR: Failed to bind queue '%s': %v", q.Name, err) // Add this logging
		return nil, amqp.Queue{}, err
	}
	log.Printf("DEBUG: Queue '%s' bound successfully.", q.Name) // Add this

	return ch, q, nil
}
