package main

import (
	"fmt"
	"log"
	// "os"
	// "os/signal"

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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	queueName := routing.PauseKey + "." + username
	_, _, err = pubsub.DeclareAndBind(conn, 
								routing.ExchangePerilDirect, 
								queueName, 
								routing.PauseKey, 
								pubsub.Transient,
							)
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	gs := gamelogic.NewGameState(username)
	
	clientGameLoop(gs)

	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan

	fmt.Println("Client connection closed...")
}

func clientGameLoop(gs *gamelogic.GameState) {
	gamelogic.PrintClientHelp()
	for {
		input := gamelogic.GetInput()
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
