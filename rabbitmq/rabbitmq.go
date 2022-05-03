package rabbitmq

import (
	"encoding/json"
	"github.com/XiovV/selly-server/models"
	"github.com/streadway/amqp"
	"os"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

func New() (*RabbitMQ, error) {
	conn, err := amqp.Dial(os.Getenv("AMQP_URL"))
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare("messages", "fanout", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(q.Name, "", "messages", false, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{conn: conn, ch: ch, q: q}, nil
}

func (r *RabbitMQ) Publish(message models.Message) error {
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return r.ch.Publish(
		"messages", "", false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        msg,
		},
	)
}

func (r *RabbitMQ) Consume() (<-chan amqp.Delivery, error) {
	return r.ch.Consume(r.q.Name, "", false, false, false, false, nil)
}
