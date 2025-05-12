package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func createDefaultKubelet(name string, roles string) *object.Kubelet {
	return &object.Kubelet{
		Config: object.KubeletConfig{
			ApiServerAddress: "http://localhost:8080",
			Name:             name,
			Roles:            roles,
			Version:          "v1",
			NodeIP:           "localhost",
		},
		Status:         "ready",
		StartTime:      time.Now(),
		Runtime:        0,
		Pods:           []object.Pod{},
		LastUpdateTime: time.Now(),
	}
}

func main() {
	// Create a test Kubelet instance.
	kube1 := createDefaultKubelet("kube1", "master")
	kube2 := createDefaultKubelet("kube2", "worker")
	kube3 := createDefaultKubelet("kube3", "worker")

	// Post the Kubelet instances to the API server to register them.
	data1, err := json.Marshal(kube1)
	if err != nil {
		panic(err)
	}

	resp1, err := http.Post(
		"http://localhost:8080/kubelet/register",
		"application/json",
		bytes.NewBuffer(data1),
	)
	if err != nil {
		panic(err)
	}

	if err := resp1.Body.Close(); err != nil {
		fmt.Printf("Error closing response body for resp1: %v\n", err)
	}

	data2, err := json.Marshal(kube2)
	if err != nil {
		panic(err)
	}

	resp2, err := http.Post(
		"http://localhost:8080/kubelet/register",
		"application/json",
		bytes.NewBuffer(data2),
	)
	if err != nil {
		panic(err)
	}

	if err := resp2.Body.Close(); err != nil {
		fmt.Printf("Error closing response body for resp2: %v\n", err)
	}

	data3, err := json.Marshal(kube3)
	if err != nil {
		panic(err)
	}

	resp3, err := http.Post(
		"http://localhost:8080/kubelet/register",
		"application/json",
		bytes.NewBuffer(data3),
	)
	if err != nil {
		panic(err)
	}

	if err := resp3.Body.Close(); err != nil {
		fmt.Printf("Error closing response body for resp3: %v\n", err)
	}

	// Get the list of Kubelet instances from the API server.
	resp, err := http.Get("http://localhost:8080/getNodes")
	if err != nil {
		panic(err)
	}

	// Print the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Response from API server: ", string(body))

	// Close the response body.
	if err := resp.Body.Close(); err != nil {
		fmt.Printf("Error closing response body: %v\n", err)
	}
}
