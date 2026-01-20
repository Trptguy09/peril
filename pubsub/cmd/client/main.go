package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	connStr := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}

	defer conn.Close()

	exchange := routing.ExchangePerilDirect
	queueName := routing.PauseKey + "." + username
	routingKey := routing.PauseKey
	queueType := amqp.Transient
	ch, q, err := pubsub.DeclareAndBind(conn, exchange, queueName, routingKey, pubsub.SimpleQueueType(queueType))
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	defer ch.Close()

	fmt.Printf("\nConnection to %v was successful.\nPress Ctr+C to stop.", q.Name)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Printf("\nRecieved signal: %s. \nShutting down...", sig)

}
