package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queuetype SimpleQueueType,
	handler func(T) AckType,
) error {

	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queuetype)
	if err != nil {
		return err
	}
	newchan, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		defer ch.Close()
		for message := range newchan {
			var target T
			if err := json.Unmarshal(message.Body, &target); err != nil {
				fmt.Println("failed to unmarshal message:", err)
				continue
			}
			ack := handler(target)
			switch ack {
			case Ack:
				fmt.Println("Ack")
				message.Ack(false)
			case NackRequeue:
				fmt.Println("NackRequeue")
				message.Nack(false, true)
			case NackDiscard:
				fmt.Println("NackDiscard")
				message.Nack(false, false)
			}
		}
	}()
	return nil
}
