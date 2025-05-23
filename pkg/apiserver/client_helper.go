package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// `getAndUnmarshalList` is a helper function to
// GET a resource and unmarshal a double-encoded JSON list.
func (c *APIClient) getAndUnmarshalList(urlSuffix string, out any) error {
	// Comcat the base URL with the URL suffix.
	url := c.BaseURL + urlSuffix

	// Create a new HTTP GET request.
	req, err := http.NewRequest("GET", url, nil)
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
			fmt.Println("Failed to close response body: ", cerr)
		}
	}()

	// Read the response body.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if the response status code is OK (200).
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", string(bodyBytes))
	}

	// Parse the response body as a list of objects.
	if err := json.Unmarshal(bodyBytes, out); err != nil {
		return fmt.Errorf("failed to unmarshal as list: %v", err)
	}

	return nil
}
