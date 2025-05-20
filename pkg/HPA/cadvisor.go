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

	// Decode the JSON response into a ContainerStat struct.
	var stats object.ContainerStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Return the status of the pod.
	return &stats, nil
}
