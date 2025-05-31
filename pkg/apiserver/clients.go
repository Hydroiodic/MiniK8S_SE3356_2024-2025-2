package apiserver

import (
	"net/http"
	"os"
	"strings"
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
		if envURL := os.Getenv("APISERVER_URL"); envURL != "" {
			baseURL = envURL
		} else {
			baseURL = APIServerURL
		}
	}

	// Ensure it has a protocol.
	if !strings.HasPrefix(baseURL, "http://") &&
		!strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}

	// Ensure it has a port.
	if !strings.Contains(baseURL[strings.Index(baseURL, "://")+3:], ":") {
		baseURL = baseURL + ":" + APIServerPort
	}

	// Create a new API client with the specified base URL and a default HTTP client.
	return &APIClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}
