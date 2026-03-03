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
	const rabbitmqConnectionString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitmqConnectionString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %s\n", err.Error())
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!.")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	gs := gamelogic.NewGameState(username)
	pauseQueueName := routing.PauseKey + "." + username
	err = pubsub.SubscribeJSON(conn, 
								routing.ExchangePerilDirect, 
								pauseQueueName, 
								routing.PauseKey, 
								pubsub.Transient,
								HandlerPause(gs),
							)
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}


	armyMovesQueueName := routing.ArmyMovesPrefix + "." + username
	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilTopic,
		armyMovesQueueName,
		routing.ArmyMovesPrefix + ".*",
		pubsub.Transient,
		HandlerMove(gs, ch),
	)
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	warQueueName := routing.WarRecognitionsPrefix 
	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilTopic,
		warQueueName,
		routing.WarRecognitionsPrefix + ".*",
		pubsub.Durable,
		HandlerWar(gs),
	)
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}


	clientGameLoop(gs, ch, username)

	fmt.Println("Client connection closed...")
}

func clientGameLoop(gs *gamelogic.GameState, ch *amqp.Channel, username string) {
	gamelogic.PrintClientHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			err := gs.CommandSpawn(input)
			if err != nil {
				log.Printf("invalid: %v\n", err.Error())
				continue
			}
		case "move":
			mv, err := gs.CommandMove(input)
			if err != nil {
				log.Printf("invalid: %v\n", err.Error())
				continue
			}
			log.Printf("Moved %v unit(s) to %v\n", len(mv.Units), mv.ToLocation)
			
			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.ArmyMovesPrefix + "." + username, mv)
			if err != nil {
				log.Printf("invalid: %v\n", err.Error())
			}
			log.Println("Move published successfully.")
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			log.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Println("Invalid command...")
		}
	}
}
