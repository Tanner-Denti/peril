package main


import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandlerPause(gs *gamelogic.GameState) func(routing.PlayingState) (pubsub.AckType) {
	return func(ps routing.PlayingState) (pubsub.AckType) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)

		return pubsub.Ack
	}
}

func HandlerMove (gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) (pubsub.AckType) {
	return func(mv gamelogic.ArmyMove) (pubsub.AckType) {
		defer fmt.Print("> ")
		mo := gs.HandleMove(mv)

		if mo == gamelogic.MoveOutComeSafe {
			return pubsub.Ack
		} else if mo == gamelogic.MoveOutcomeMakeWar {
			err := pubsub.PublishJSON(ch, 
						routing.ExchangePerilTopic, 
						routing.WarRecognitionsPrefix + "." + gs.Player.Username, 
						gamelogic.RecognitionOfWar{
							Attacker: mv.Player,
							Defender: gs.GetPlayerSnap(),
						},
					)
			if err != nil {
				log.Println("Transient error in HandlerMove, requeueing...")
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		
		return pubsub.NackDiscard
	}
}

func HandlerWar (gs *gamelogic.GameState) func(rw gamelogic.RecognitionOfWar) (pubsub.AckType) {
	return func(rw gamelogic.RecognitionOfWar) (pubsub.AckType) {
		defer fmt.Print("> ")
		outcome, _, _:= gs.HandleWar(rw)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		default:
			log.Println("Error, no valid war outcome...")
			return pubsub.NackDiscard
		}
	}
}