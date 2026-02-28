package main

import( 
	"fmt"
	"log"
	"os"
	"os/signal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const rabbitmqConnectionString = "amqp://guest:guest@localhost:5672/"
	amqpConn, err := amqp.Dial(rabbitmqConnectionString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %s\n", err.Error())
	}
	defer amqpConn.Close()

	fmt.Println("Peril game server connected to RabbitMQ!.")
	
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	
	fmt.Println("\nRabbitMQ connection closed.")
}
