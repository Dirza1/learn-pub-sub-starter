package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonByte, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "aplication/json",
		Body:        jsonByte})
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
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	queChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	var durable bool
	var autoDelete bool
	var exclusive bool
	switch queueType {
	case Durable:
		durable = true
		autoDelete = false
		exclusive = false
	case Transient:
		durable = false
		autoDelete = true
		exclusive = true
	}
	playerQue, err := queChan.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = queChan.QueueBind(playerQue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return queChan, playerQue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) (amqp.Acknowledger, error) {

	chanCheck, queCheck, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return nil, err
	}

	deliveryChann, err := chanCheck.Consume(queCheck.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	go func() {
		for val := range deliveryChann {
			var data T
			if uerr := json.Unmarshal(val.Body, &data); uerr != nil {
				fmt.Printf("Err: %v\n", uerr)
				continue
			}
			handler(data)
			if aerr := val.Ack(false); aerr != nil {
				fmt.Printf("Ack error: %v\n", aerr)
				continue
			}

		}
	}()
	return nil, nil

}
