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
	// If baseURL is empty, set it to the default API server URL.
	if baseURL == "" {
		baseURL = os.Getenv("APISERVER_URL")
	}
	// If the environment variable is also not set, use the default value.
	if baseURL == "" {
		baseURL = APIServerUrl
	}

	// Create a new API client with the specified base URL and a default HTTP client.
	return &APIClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}
