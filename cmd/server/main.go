package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const rabbitmqConnectionString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitmqConnectionString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %s\n", err.Error())
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!.")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}

	_, queue, err := pubsub.DeclareAndBind(
					conn, 
					routing.ExchangePerilTopic,
					routing.GameLogSlug,
					routing.GameLogSlug + ".*",
					pubsub.Durable,
				)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound !\n", queue.Name)

	serverGameLoop(ch)
	
	fmt.Println("\nRabbitMQ connection closed.")
}

func serverGameLoop(ch *amqp.Channel) {
	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			log.Println("Sending pause message")
			err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{ IsPaused: true })
			if err != nil {
				log.Fatalf("error: %s\n", err.Error())
			}
		case "resume":
			log.Println("Sending resume message")
			err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{ IsPaused: false })
			if err != nil {
				log.Fatalf("error: %s\n", err.Error())
			}
		case "quit":
			log.Println("Exiting game...")
			return
		default:
			log.Println("Invalid command")
		}
	}
}