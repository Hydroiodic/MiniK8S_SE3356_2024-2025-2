package mqtemplate

import (
	"encoding/json"
	"fmt"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func CreatePodMessage(pod object.Pod) (string, error) {
	// Create a message for creating a pod.
	jsonData, err := json.Marshal(pod)
	if err != nil {
		fmt.Println("Failed to marshal pod: ", err)
		return "", err
	}

	return string(jsonData), nil
}

func SendMessageToQueue(queueName string, body string) error {
	// Create connection to RabbitMQ server.
	conn, err := Connect(RabbitMQUrl)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ: ", err)
		return err
	}

	// Ensure the connection is closed when the function exits.
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("Error closing connection: ", err)
		}
	}()

	// Create a channel.
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel: ", err)
		return err
	}

	// Ensure the channel is closed when the function exits.
	defer func() {
		if err := ch.Close(); err != nil {
			fmt.Println("Error closing channel: ", err)
		}
	}()

	// Create a queue.
	err = CreateQueue(ch, queueName)
	if err != nil {
		fmt.Println("Failed to declare a queue: ", err)
		return err
	}

	// Send a message to the queue.
	if err := SendMessage(ch, queueName, body); err != nil {
		fmt.Println("Failed to publish a message: ", err)
		return err
	}

	return nil
}

func ConsumeMessageOnQueue(
	queueName string,
	handler func(msg map[string]any) error,
) error {
	// Create connection to RabbitMQ server.
	conn, err := Connect(RabbitMQUrl)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ: ", err)
		return err
	}

	// Ensure the connection is closed when the function exits.
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("Error closing connection: ", err)
		}
	}()

	// Create a channel.
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel: ", err)
		return err
	}

	// Ensure the channel is closed when the function exits.
	defer func() {
		if err := ch.Close(); err != nil {
			fmt.Println("Error closing channel: ", err)
		}
	}()

	// Ensure the queue exists.
	err = CreateQueue(ch, queueName)
	if err != nil {
		fmt.Println("Failed to declare a queue: ", err)
		return err
	}

	// Consume messages from the queue.
	msgs, err := ConsumeMessages(ch, queueName)
	if err != nil {
		fmt.Println("Failed to register a consumer: ", err)
		return err
	}

	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")

	// Create an infinite loop to keep the program running.
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			// Print the message.
			fmt.Println("Received a message: ", string(d.Body))

			// Consume the message and handle it.
			var msg map[string]interface{}
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				fmt.Println("Failed to unmarshal message: ", err)
				continue
			}

			// Call the handler function with the message.
			if err := handler(msg); err != nil {
				fmt.Println("Consumer handler error: ", err)
			}
		}
	}()

	<-forever

	return nil
}
