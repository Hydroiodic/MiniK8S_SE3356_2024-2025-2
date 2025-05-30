package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

// TODO: the function is low in efficiency.
// FIXME: namespace should be considered here.
func (c *APIClient) GetReplicasetyName(
	replicasetName string,
) (object.ReplicaSet, error) {
	replicasets, err := c.GetReplicasets()
	if err != nil {
		return object.ReplicaSet{}, err
	}

	for _, replicaset := range replicasets {
		if replicaset.Metadata.Name == replicasetName {
			return replicaset, err
		}
	}

	return object.ReplicaSet{}, fmt.Errorf(
		"replicaset %s not found",
		replicasetName,
	)
}

func (c *APIClient) GetReplicasets() ([]object.ReplicaSet, error) {
	// Construct the URL for the Kubelet get nodes endpoint.
	url := c.BaseURL + ReplicasetGetURL

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
	var replicasets []object.ReplicaSet
	if err := json.NewDecoder(resp.Body).Decode(&replicasets); err != nil {
		// If parsing fails, return an error.
		return nil, err
	}

	return replicasets, nil
}

func (c *APIClient) CreateReplicaset(replicaset *object.ReplicaSet) error {
	// Construct the URL for the Pod creation endpoint.
	url := c.BaseURL + ReplicasetCreateURL

	// Convert the Pod object to JSON to be sent in the request body.
	replicasetJSON, err := json.Marshal(replicaset)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(replicasetJSON))
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

func (c *APIClient) DeleteReplicaset(rsName string) error {
	rs, err := c.GetReplicasetyName(rsName)
	if err != nil {
		return err
	}

	url := c.BaseURL + ReplicasetDeleteURL
	pods, err := c.GetPods()

	if err != nil {
		return err
	}

	var matchPods []object.Pod

	rsJSON, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(rsJSON))
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

	for _, p := range pods {
		if HasMatchingLabels(
			rs.Spec.Template.Metadata.Labels,
			p.Metadata.Labels,
		) {
			matchPods = append(matchPods, p)
		}
	}

	for _, p := range matchPods {
		err = c.DeletePod(&p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *APIClient) UpdateReplicaset(r *object.ReplicaSet) error {
	// Construct the URL for the Kubelet get nodes endpoint.
	url := c.BaseURL + ReplicasetUpdateURL

	replicasetJSON, err := json.Marshal(r)
	if err != nil {
		return err
	}

	// Create a new HTTP GET request.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(replicasetJSON))
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

	// Parse the response body as a list of Replicaset objects.
	var replicasets []object.ReplicaSet
	if err := json.NewDecoder(resp.Body).Decode(&replicasets); err != nil {
		// If parsing fails, return an error.
		return err
	}

	return nil
}
