package apiserver

import (
	"net/http"
	"os"
)

type APIClient struct {
	// The base URL for the API server.
	BaseURL string
	// The HTTP client used to make requests.
	Client *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	// If baseURL is empty, get it from the environment variable or use the default.
	if baseURL == "" {
		// Check if the environment variable APISERVER_URL is set.
		if envURL := os.Getenv("APISERVER_URL"); envURL != "" {
			baseURL = envURL
		} else {
			baseURL = APIServerURL
		}

		// Add the port to the base URL.
		baseURL = baseURL + ":" + APIServerPort
	}

	// Create a new API client with the specified base URL and a default HTTP client.
	return &APIClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}
