package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	body, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ctx := context.Background()

	err = ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		return err
	}
	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(val)
	if err != nil {
		return err
	}

	body := b.Bytes()
	ctx := context.Background()

	err = ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        body,
	})
	if err != nil {
		return err
	}
	return nil
}
