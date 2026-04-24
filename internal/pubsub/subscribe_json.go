package pubsub

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {

	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	amqpChan, err := ch.Consume(
		queue.Name, // queue
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return err
	}

	go func() {

		for d := range amqpChan {
			var msg T
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				d.Nack(false, false)
				continue
			}

			handler(msg)
			d.Ack(false)
		}

	}()

	return nil

}
