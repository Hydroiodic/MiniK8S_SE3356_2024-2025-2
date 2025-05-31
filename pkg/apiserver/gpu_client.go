package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func (c *APIClient) CreateGPUJob(job object.Job) error {
	// Construct the URL for the Pod creation endpoint.
	url := c.BaseURL + GPUJobsCreateURL

	// Convert the Pod object to JSON to be sent in the request body.
	GPUJobJSON, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(GPUJobJSON))
	if err != nil {
		return err
	}

	fmt.Println(job)

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

func (c *APIClient) GetAllGPUJob() ([]object.Job, error) {
	var jobs []object.Job
	err := c.getAndUnmarshalList(GPUJobsGetURL, &jobs)

	return jobs, err
}
