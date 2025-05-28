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

func (c *APIClient) AddDNS(dns *object.DNS) error {
	// Construct the URL for the DNS addition endpoint.
	url := c.BaseURL + DNSAddURL

	// Convert the DNS object to JSON to be sent in the request body.
	dnsJSON, err := json.Marshal(dns)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the DNS as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dnsJSON))
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

func (c *APIClient) DeleteDNS(dns *object.DNS) error {
	// Construct the URL for the DNS deletion endpoint.
	url := c.BaseURL + DNSDeleteURL

	// Convert the DNS object to JSON to be sent in the request body.
	dnsJSON, err := json.Marshal(dns)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the DNS as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dnsJSON))
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

func (c *APIClient) DeleteDNSByName(name, namespace string) error {
	// Create a DNS object with the provided name and namespace.
	dns := &object.DNS{
		Metadata: object.Metadata{
			Name:      name,
			Namespace: namespace,
		},
	}

	// Call the DeleteDNS method to delete the DNS configuration.
	return c.DeleteDNS(dns)
}

// This function will return all DNS parsing and reverse proxy information.
func (c *APIClient) GetForwardingInfo() (*object.ForwardingInfo, error) {
	// Construct the URL for the DNS resolve information endpoint.
	url := c.BaseURL + DNSGetForwardingInfoURL

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

	// Parse the response body as a ForwardingInfo object.
	var info object.ForwardingInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		// If parsing fails, return an error.
		return nil, err
	}

	return &info, nil
}
