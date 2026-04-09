package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func subscribe[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) Acktype, unmarshaller func([]byte) (T, error)) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	defer func() {
		if err := ch.Close(); err != nil {
			fmt.Printf("error closing channel: %v\n", err)
		}
	}()
	err = ch.Qos(10, 0, false)
	if err != nil {
		return err
	}
	deliveries, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			t, err := unmarshaller([]byte(d.Body))
			if err != nil {
				fmt.Printf("error unmarshalling message: %v\n", err)
				continue
			}
			switch handler(t) {
			case Ack:
				d.Ack(false)
			case NackDiscard:
				d.Nack(false, false)
			case NackRequeue:
				d.Nack(false, true)
			}
		}
	}()
	return nil
}
