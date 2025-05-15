package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"

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

	// Create a new API client to interact with the API server.
	client := apiserver.NewAPIClient("")

	// Get the list of nodes from the API server.
	nodes, err := client.GetNodes()
	if err != nil {
		return err
	}

	// There should be at least one node available to assign the pod.
	if len(nodes) == 0 {
		return fmt.Errorf("no nodes available to assign the pod")
	}

	// Use crypto/rand to securely select a random node index
	max := big.NewInt(int64(len(nodes)))
	nBig, err := rand.Int(rand.Reader, max)

	// Check for errors in generating the random number.
	if err != nil {
		return fmt.Errorf("failed to generate secure random index: %v", err)
	}

	// Select a random node from the list of nodes.
	node := &nodes[nBig.Int64()]

	// Assign the pod to a node.
	err = client.AssignPodToNode(&pod, node.Config.Name)
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
