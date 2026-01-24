package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(state routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(state)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		outcome := gs.HandleMove(move)

		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			})
			if err != nil {
				fmt.Printf("could not publish war recognition: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(msg gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, winner, loser := gs.HandleWar(msg)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			if err := publishGameLog(ch, gs.GetUsername(), fmt.Sprintf("%s won a war against %s", winner, loser)); err != nil {
				fmt.Printf("could not publish game log: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			if err := publishGameLog(ch, gs.GetUsername(), fmt.Sprintf("%s won a war against %s", winner, loser)); err != nil {
				fmt.Printf("could not publish game log: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			if err := publishGameLog(ch, gs.GetUsername(), fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)); err != nil {
				fmt.Printf("could not publish game log: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Printf("unknown war outcome: %v", outcome)
			return pubsub.NackDiscard
		}
	}
}

func publishGameLog(ch *amqp.Channel, username, logMsg string) error {
	logRk := routing.GameLogSlug + "." + username

	logMsgStruct := routing.GameLog{
		Username:    username,
		Message:     logMsg,
		CurrentTime: time.Now(),
	}
	return pubsub.PublishGob(ch, routing.ExchangePerilTopic, logRk, logMsgStruct)
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

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, routing.ArmyMovesPrefix+".*", pubsub.SimpleQueueType(pubsub.QueueTransient), handlerMove(gs, pubchan))
	if err != nil {
		log.Fatalf("unable to bind army_moves to user")
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "war", routing.WarRecognitionsPrefix+".*", pubsub.SimpleQueueType(pubsub.QueueDurable), handlerWar(gs, pubchan))
	if err != nil {
		log.Fatalf("unable to bind war_recognitions to user")
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
