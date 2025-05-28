package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func (c *APIClient) RegisterKubelet(kubelet *object.Kubelet) error {
	// Construct the URL for the Kubelet registration endpoint.
	url := c.BaseURL + KubeletRegisterURL

	// Convert the Kubelet object to JSON to be sent in the request body.
	kubeletJSON, err := json.Marshal(kubelet)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the Kubelet JSON as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(kubeletJSON))
	if err != nil {
		return err
	}

	// Send the request and return the response.
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	// Ensure the response body is closed after use.
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Log the error if closing the response body fails.
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	// Check if the response status code is OK (200).
	if resp.StatusCode != http.StatusOK {
		// If not, read the response body and return an error.
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", string(bodyBytes))
	}

	return nil
}

func (c *APIClient) HeartbeatKubelet(kubelet *object.Kubelet) error {
	// Construct the URL for the Kubelet heartbeat endpoint.
	url := c.BaseURL + KubeletHeartbeatURL

	kubletBytes, err := json.Marshal(kubelet)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the node name as the body.
	req, err := http.NewRequest("POST", url, bytes.NewReader(kubletBytes))
	if err != nil {
		return err
	}
	// Set the content type to JSON.
	req.Header.Set("Content-Type", "application/json")

	// Send the request and return the response.
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	// Ensure the response body is closed after use.
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Log the error if closing the response body fails.
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	// Check if the response status code is OK (200).
	if resp.StatusCode != http.StatusOK {
		// If not, read the response body and return an error.
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", string(bodyBytes))
	}

	return nil
}
