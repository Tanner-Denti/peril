package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Client connection closed...")
}
