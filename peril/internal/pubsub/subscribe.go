package pubsub

import (
	"bytes"
	"encoding/gob"
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

func helperSubscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queuetype SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte, *T) error,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queuetype)
	if err != nil {
		return err
	}
	err = ch.Qos(10, 0, false)
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
			if err := unmarshaller(message.Body, &target); err != nil {
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

func jsonUnmarshal[T any](data []byte, target *T) error {
	return json.Unmarshal(data, target)
}

func gobUnmarshal[T any](data []byte, target *T) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(target)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queuetype SimpleQueueType,
	handler func(T) AckType,
) error {

	return helperSubscribe[T](conn, exchange, queueName, key, queuetype, handler, gobUnmarshal[T])
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queuetype SimpleQueueType,
	handler func(T) AckType,
) error {

	return helperSubscribe[T](conn, exchange, queueName, key, queuetype, handler, jsonUnmarshal[T])
}
