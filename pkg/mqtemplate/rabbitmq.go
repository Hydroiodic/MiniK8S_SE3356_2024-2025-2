package mqtemplate

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// Connect to RabbitMQ server.
func Connect(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Create RabbitMQ channel.
func CreateQueue(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	return err
}

// Send message to a specific RabbitMQ queue.
func SendMessage(ch *amqp.Channel, queueName string, body string) error {
	err := ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

	return err
}

// Consume messages from a specific RabbitMQ queue.
func ConsumeMessages(
	ch *amqp.Channel,
	queueName string,
) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	return msgs, err
}
