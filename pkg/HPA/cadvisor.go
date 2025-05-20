package hpa

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type CAdvisorClient struct {
	BaseURL string
	Client  *http.Client
}

func NewCAdvisorClient(host string, port int) *CAdvisorClient {
	// If host or port is not specified, use the default ones
	if host == "" {
		host = DefaultCadvisorHost
	}

	if port <= 0 {
		port = DefaultCadvisorPort
	}

	return &CAdvisorClient{
		BaseURL: fmt.Sprintf("http://%s:%d", host, port),
		Client:  &http.Client{},
	}
}

// Get the CPU/Memory usage of a specific pod
func (c *CAdvisorClient) ContainerStats(
	podName string,
) (*object.ContainerStats, error) {
	// Construct the URL for the pod metrics and make the GET request.
	resp, err := http.Get(
		fmt.Sprintf("%s%s%s", c.BaseURL, CadvisorEndpoint, podName),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod stats: %v", err)
	}

	// Ensure the response body is closed after reading.
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body: ", err)
		}
	}()

	// Check if the response status code is OK (200).
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get pod metrics: %s", resp.Status)
	}

	// Parse as a map of strings to interface{}.
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Get the first value from the map (that is what we expect).
	var containerStats object.ContainerStats

	for _, v := range result {
		// Marshal the data to JSON.
		dataBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal container data: %v", err)
		}

		// Unmarshal the JSON data into the ContainerStats struct.
		if err := json.Unmarshal(dataBytes, &containerStats); err != nil {
			return nil, fmt.Errorf(
				"failed to unmarshal container data: %v",
				err,
			)
		}

		break
	}

	// Return the status of the container.
	return &containerStats, nil
}
