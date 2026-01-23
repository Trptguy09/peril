package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(state routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(state)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(move)
	}
}

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

	pubchan, err := conn.Channel()
	if err != nil {
		fmt.Printf("%v", err)
	}

	defer pubchan.Close()
	defer conn.Close()

	queueName := routing.PauseKey + "." + username

	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.SimpleQueueType(pubsub.QueueTransient), handlerPause(gs))

	if err != nil {
		log.Fatalf("could not subscribe: %v", err)
	}

	rk := "army_moves.*"
	qn := "army_moves." + username
	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, qn, rk, pubsub.QueueTransient)
	if err != nil {
		log.Fatalf("unable to bind army_moves to user")
	}

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

			moveRK := "army_moves." + gs.GetUsername()
			move, err := gs.CommandMove(words)
			if err != nil {
				fmt.Printf("%v", err)
				continue
			}

			err = pubsub.PublishJSON(
				pubchan,
				routing.ExchangePerilTopic,
				moveRK,
				move)
			if err != nil {
				fmt.Printf("%v", err)
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
