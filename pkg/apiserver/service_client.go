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

func (c *APIClient) GetServices() ([]object.Service, error) {
	var svcs []object.Service
	err := c.getAndUnmarshalList(ServiceGetURL, &svcs)

	return svcs, err
}

func (c *APIClient) CreateService(svc *object.Service) error {
	// Construct the URL for the Service creation endpoint.
	url := c.BaseURL + ServiceCreateURL

	// Convert the Service object to JSON to be sent in the request body.
	svcJSON, err := json.Marshal(svc)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the service as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(svcJSON))
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

func (c *APIClient) DeleteService(svc *object.Service) error {
	// Construct the URL for the Service deletion endpoint.
	url := c.BaseURL + ServiceDeleteURL

	// Convert the Service object to JSON to be sent in the request body.
	svcJSON, err := json.Marshal(svc)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the service as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(svcJSON))
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
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	// Check if the response status code is OK (200).
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", string(bodyBytes))
	}

	return nil
}

func (c *APIClient) DeleteServiceByName(name, namespace string) error {
	// Construct a service object with the provided namespace and name.
	svc := &object.Service{
		Metadata: object.Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}

	// Call the DeleteService method to delete the service.
	return c.DeleteService(svc)
}
