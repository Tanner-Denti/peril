package main


import (
	"fmt"
	"log"
	"time"

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

func HandlerWar (gs *gamelogic.GameState, ch *amqp.Channel) func(rw gamelogic.RecognitionOfWar) (pubsub.AckType) {
	return func(rw gamelogic.RecognitionOfWar) (pubsub.AckType) {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		gameLog := routing.GameLog {
			CurrentTime: time.Now(),
			Message: "",
			Username: gs.GetUsername(),
		}
		
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeYouWon:
			gameLog.Message = fmt.Sprintf("%s won a war against %s", winner, loser)
			err := pubsub.PublishGameLog(ch, gameLog)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack		
		case gamelogic.WarOutcomeOpponentWon:
			gameLog.Message = fmt.Sprintf("%s won a war against %s", winner, loser)
			err := pubsub.PublishGameLog(ch, gameLog)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			gameLog.Message = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			err := pubsub.PublishGameLog(ch, gameLog)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			log.Println("Error, no valid war outcome...")
			return pubsub.NackDiscard
		}
	}
}

