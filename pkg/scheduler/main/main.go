package main

import (
	"bytes"
	"encoding/json"
	"net/http"

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

	// Assign the pod to a node by sending a POST request to APIServer.
	resp, err := http.Post(
		apiserver.APIServerUrl+"/pod/assignNodetoPod",
		"application/json",
		bytes.NewBuffer(podjson),
	)
	if err != nil {
		return err
	}

	// Check the response status code.
	if err := resp.Body.Close(); err != nil {
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
