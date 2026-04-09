package pubsub

import (
	"bytes"
	"encoding/gob"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) Acktype) error {
	return subscribe[T](conn, exchange, queueName, key, queueType, handler, func(data []byte) (T, error) {
		var t T
		buffer := bytes.NewBuffer(data)
		decoder := gob.NewDecoder(buffer)
		return t, decoder.Decode(&t)
	})
}
