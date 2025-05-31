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

func (c *APIClient) GetPods() ([]object.Pod, error) {
	var pods []object.Pod
	err := c.getAndUnmarshalList(PodGetURL, &pods)

	return pods, err
}

func (c *APIClient) CreatePod(pod *object.Pod) error {
	// Construct the URL for the Pod creation endpoint.
	url := c.BaseURL + PodCreateURL

	// Convert the Pod object to JSON to be sent in the request body.
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(podJSON))
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

func (c *APIClient) DeletePod(pod *object.Pod) error {
	// Construct the URL for the Pod deletion endpoint.
	url := c.BaseURL + PodDeleteURL

	// Convert the Pod object to JSON to be sent in the request body.
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(podJSON))
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
			fmt.Println("Failed to close response body: ", cerr)
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

func (c *APIClient) DeletePodByName(name, namespace string) error {
	// Construct the URL for the Pod deletion endpoint.
	pod := &object.Pod{
		Metadata: object.Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}

	// Call the DeletePod method with the constructed pod object.
	return c.DeletePod(pod)
}

func (c *APIClient) AssignPodToNode(pod *object.Pod, nodeName string) error {
	// Construct the URL for the Pod assignment endpoint.
	url := c.BaseURL + PodAssignURL

	// Convert the Pod object to JSON to be sent in the request body.
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	// `nodeName` is passed as a query parameter.
	req, err := http.NewRequest(
		"POST",
		url+"?nodeName="+nodeName,
		bytes.NewBuffer(podJSON),
	)
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
