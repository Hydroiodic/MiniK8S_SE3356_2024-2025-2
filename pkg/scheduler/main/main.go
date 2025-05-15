package main

import (
	"encoding/json"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func handleCreateNewPod(msg map[string]interface{}) error {
	// Handle the message for creating a new pod.
	podjson, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Unmarshal the JSON message into a Pod struct.
	var pod object.Pod
	if err := json.Unmarshal(podjson, &pod); err != nil {
		return err
	}

	// TODO: Get all nodes and choose one randomly.

	// Create a new API client to interact with the API server.
	client := apiserver.NewAPIClient("")

	// Assign the pod to a node.
	err = client.AssignPodToNode(&pod)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Consume messages from the RabbitMQ queue for creating pods.
	err := mqtemplate.ConsumeMessageOnQueue(
		mqtemplate.CreatePodQueueName, handleCreateNewPod,
	)
	if err != nil {
		// Log the error and exit the program
		panic(err)
	}

	// Block the main goroutine to keep the program running.
	// This is a simple way to keep the program running indefinitely.
	select {}
}
