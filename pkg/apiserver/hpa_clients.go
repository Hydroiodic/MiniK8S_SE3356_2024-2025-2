package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func (c *APIClient) GetHpas() ([]object.HorizontalPodAutoscaler, error) {
	// Construct the URL for the Kubelet get nodes endpoint.
	url := c.BaseURL + HpaGetURL

	// Create a new HTTP GET request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Send the request and return the response.
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("%s", string(bodyBytes))
	}

	// Parse the response body as a list of Replicaset objects.
	var hpas []object.HorizontalPodAutoscaler
	if err := json.NewDecoder(resp.Body).Decode(&hpas); err != nil {
		// If parsing fails, return an error.
		return nil, err
	}

	return hpas, nil
}

func (c *APIClient) GeHpayName(
	hpaName string,
	namespace string,
) (object.HorizontalPodAutoscaler, error) {
	hpas, err := c.GetHpas()
	if err != nil {
		return object.HorizontalPodAutoscaler{}, err
	}

	for _, hpa := range hpas {
		if hpa.Metadata.Name == hpaName &&
			hpa.Metadata.Namespace == namespace {
			return hpa, err
		}
	}

	return object.HorizontalPodAutoscaler{}, fmt.Errorf(
		"not HPA named" + hpaName + "found",
	)
}

func (c *APIClient) CreateHpa(hpa *object.HorizontalPodAutoscaler) error {
	// Construct the URL for the Pod creation endpoint.
	url := c.BaseURL + HpaCreateURL

	// Convert the Pod object to JSON to be sent in the request body.
	hpaJSON, err := json.Marshal(hpa)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(hpaJSON))
	if err != nil {
		return err
	}

	fmt.Println(hpa)

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

func (c *APIClient) DeleteHpa(pod *object.HorizontalPodAutoscaler) error {
	// Construct the URL for the Pod deletion endpoint.
	url := c.BaseURL + HpaDeleteURL

	// Convert the Pod object to JSON to be sent in the request body.
	HpaJSON, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(HpaJSON))
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
