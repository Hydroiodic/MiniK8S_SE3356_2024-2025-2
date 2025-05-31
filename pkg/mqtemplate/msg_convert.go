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

func CreateServiceMessage(service object.Service) (string, error) {
	// Create a message for creating a service.
	jsonData, err := json.Marshal(service)
	if err != nil {
		fmt.Println("Failed to marshal service: ", err)
		return "", err
	}

	return string(jsonData), nil
}
