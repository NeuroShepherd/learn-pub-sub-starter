package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	QueueTransient SimpleQueueType = iota
	QueueDurable
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable := queueType == QueueDurable
	autoDelete := queueType == QueueTransient
	exclusive := queueType == QueueTransient

	queue, err := ch.QueueDeclare(
		queueName,
		durable,    // durable
		autoDelete, // delete when unused
		exclusive,  // exclusive
		false,      // no-wait
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(
		queue.Name, // queue name
		key,        // routing key
		exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, err
	}

	return ch, queue, nil
}
