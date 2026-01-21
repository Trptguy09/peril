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
	_, _, err = pubsub.DeclareAndBind(conn, exchange, queueName, routingKey, pubsub.SimpleQueueType(queueType))
	if err != nil {
		log.Fatalf("%v", err)
		return
	}
	gs := gamelogic.NewGameState(username)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		cmd := words[0]

		if cmd == "spawn" {

			before := gs.GetPlayerSnap()

			err = gs.CommandSpawn(words)
			if err != nil {
				fmt.Printf("%v", err)
				continue
			}

			after := gs.GetPlayerSnap()

			var newUnit gamelogic.Unit

			for _, u := range after.Units {

				found := false

				for _, old := range before.Units {

					if u.ID == old.ID {
						found = true
						break
					}
				}

				if !found {
					newUnit = u
					break
				}
			}

			fmt.Printf("%v", newUnit.ID)

		} else if cmd == "move" {

			_, err := gs.CommandMove(words)
			if err != nil {
				fmt.Printf("%v", err)
				continue
			}

			fmt.Printf("Move Successful")

		} else if cmd == "status" {

			gs.CommandStatus()

		} else if cmd == "help" {

			gamelogic.PrintClientHelp()

		} else if cmd == "spam" {

			fmt.Printf("Spamming not allowed yet!")

		} else if cmd == "quit" {

			gamelogic.PrintQuit()
			break
		} else {
			fmt.Println("cmd unknown")
			continue
		}
	}
}
