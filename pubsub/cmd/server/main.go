package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	connStr := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	defer conn.Close()
	defer ch.Close()

	fmt.Println("Connection was successful. Press Ctr+C to stop.")

	gamelogic.PrintServerHelp()

	for {
		inputSlice := gamelogic.GetInput()
		if len(inputSlice) == 0 {
			continue
		}
		cmd := inputSlice[0]

		if cmd == "pause" {
			fmt.Println("Sending a pause message")
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true},
			)

			if err != nil {
				log.Fatalf("%v", err)
			}
		} else if cmd == "resume" {
			fmt.Println("Sending a resume message")
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false},
			)
			if err != nil {
				log.Fatalf("%v", err)
			}
		} else if cmd == "quit" {
			fmt.Println("Exiting")
			break
		} else {
			fmt.Println("cmd unknown")
		}
	}
}
