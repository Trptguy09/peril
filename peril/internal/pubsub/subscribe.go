package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queuetype SimpleQueueType,
	handler func(T),
) error {

	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, SimpleQueueType(queuetype))
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
			var gen T
			if err := json.Unmarshal(message.Body, &gen); err != nil {
				fmt.Println("failed to unmarshal message:", err)
				continue
			}
			handler(gen)
			message.Ack(false)
		}
	}()
	return nil
}
