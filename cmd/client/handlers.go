package main

import (
	"fmt"
	"log"

	"github.com/Denisowiec/learn-pub-sub-starter/internal/gamelogic"
	"github.com/Denisowiec/learn-pub-sub-starter/internal/pubsub"
	"github.com/Denisowiec/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")

		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func HandlerMove(gs *gamelogic.GameState, pubChan *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(am gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		outcome := gs.HandleMove(am)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			qName := fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername())
			message := gamelogic.RecognitionOfWar{
				Attacker: am.Player,
				Defender: gs.GetPlayerSnap(),
			}
			err := pubsub.PublishJSON(pubChan, routing.ExchangePerilTopic, qName, message)
			if err != nil {
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

func HandlerWar(gs *gamelogic.GameState, pubChan *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(recWar gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, winner, loser := gs.HandleWar(recWar)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			msg := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := PublishGameLog(pubChan, msg, gs.GetUsername())
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			msg := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := PublishGameLog(pubChan, msg, gs.GetUsername())
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			msg := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			err := PublishGameLog(pubChan, msg, gs.GetUsername())
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			log.Println("Error when processing war outcome")
			return pubsub.NackDiscard
		}
	}
}
