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

func (c *APIClient) GetPV(
	pvName string,
) (*object.PersistentVolume, error) {
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

func (c *APIClient) GetPVC(
	namespace, pvcName string,
) (*object.PersistentVolumeClaim, error) {
	// Construct the URL for the PVC retrieval endpoint.
	url := fmt.Sprintf("%s/%s/%s", PVCGetURL, namespace, pvcName)

	// Create a new HTTP GET request.
	req, err := http.NewRequest("GET", c.BaseURL+url, nil)
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
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	// Check if the response status code is OK (200).
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", string(bodyBytes))
	}

	var pvc object.PersistentVolumeClaim
	if err := json.NewDecoder(resp.Body).Decode(&pvc); err != nil {
		return nil, fmt.Errorf("failed to decode PVC response: %v", err)
	}

	return &pvc, nil
}

func (c *APIClient) CreatePVC(
	pvc *object.PersistentVolumeClaim,
) (*object.PersistentVolumeClaim, error) {
	// Construct the URL for the PVC creation endpoint.
	url := c.BaseURL + PVClaimCreateURL

	// Marshal the PVC object to JSON.
	body, err := json.Marshal(pvc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PVC: %v", err)
	}

	// Create a new HTTP POST request with the JSON body.
	req, err := http.NewRequest(
		"POST",
		url,
		io.NopCloser(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request and return the response.
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// Ensure the response body is closed after use.
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Println("Failed to close response body: ", cerr)
		}
	}()

	// Check if the response status code is OK (201).
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", string(bodyBytes))
	}

	var createdPVC object.PersistentVolumeClaim
	if err := json.NewDecoder(resp.Body).Decode(&createdPVC); err != nil {
		return nil, fmt.Errorf("failed to decode created PVC response: %v", err)
	}

	return &createdPVC, nil
}
