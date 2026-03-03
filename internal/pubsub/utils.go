package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int
const (
	Durable SimpleQueueType = iota
	Transient 
)

type AckType int
const (
	Ack = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) (AckType),
) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveryCh, err := ch.Consume("", "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		defer ch.Close()
		for delivery := range deliveryCh {
			var msg T
			err := json.Unmarshal(delivery.Body, &msg)
			if err != nil {
				log.Printf("Error in SubscribeJSON - exchange: %v, queue: %v, errMsg: %v\n", exchange, queueName, err.Error())
			}

			ackType := handler(msg)
			switch ackType {
			case Ack:
				err = delivery.Ack(false)
				log.Println("Sent Ack")
			case NackRequeue:
				err = delivery.Nack(false, true)
				log.Println("Sent NackRequeue")
			case NackDiscard:
				err = delivery.Nack(false, false)
				log.Println("Sent NackDiscard")
			default:
				err = fmt.Errorf("Invalid ack type returned from handler.\n")
			}
			if err != nil {
				log.Printf("Error in SubscribeJSON - exchange: %v, queue: %v, errMsg: %v\n", exchange, queueName, err.Error())
			}
		}
	}()

	return nil
}

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body: jsonBytes,
	})
	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange, 
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	
	queue, err := ch.QueueDeclare(queueName, 
						queueType==Durable, 
						queueType==Transient,
						queueType==Transient,
						false,
						amqp.Table{ "x-dead-letter-exchange": "peril_dlx" },
					)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	
	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, queue, nil
}