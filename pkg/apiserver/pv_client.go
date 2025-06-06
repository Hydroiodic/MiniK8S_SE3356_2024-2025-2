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

func (c *APIClient) GetPVs() ([]object.PersistentVolume, error) {
	var nodes []object.PersistentVolume
	err := c.getAndUnmarshalList(PVGetURL, &nodes)

	return nodes, err
}

func (c *APIClient) GetPV(pvName string) (*object.PersistentVolume, error) {
	// Construct the URL for the PV retrieval endpoint.
	url := c.BaseURL + PVGetURL + "/" + pvName

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
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	// Check if the response status code is OK (200).
	if resp.StatusCode != http.StatusOK {
		// If not, read the response body and return an error.
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", string(bodyBytes))
	}

	var pv object.PersistentVolume
	if err := json.NewDecoder(resp.Body).Decode(&pv); err != nil {
		return nil, fmt.Errorf("failed to decode PV response: %v", err)
	}

	return &pv, nil
}

func (c *APIClient) CreatePV(pv *object.PersistentVolume) error {
	// Construct the URL for the PV creation endpoint.
	url := c.BaseURL + PVCreateURL
	// Marshal the PV object to JSON.
	body, err := json.Marshal(pv)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the JSON body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
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

	// Check if the response status code is Created (200).
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", string(bodyBytes))
	}

	return nil
}

func (c *APIClient) DeletePV(pv *object.PersistentVolume) error {
	// Construct the URL for the PV deletion endpoint.
	url := c.BaseURL + PVDeleteURL
	// Marshal the PV object to JSON.
	pvJSON, err := json.Marshal(pv)
	if err != nil {
		return err
	}
	// Create a new HTTP POST request with the JSON body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(pvJSON))
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

func (c *APIClient) DeletePVByName(pvName string) error {
	pv := &object.PersistentVolume{
		Metadata: object.Metadata{
			Name: pvName,
		},
	}

	return c.DeletePV(pv)
}
