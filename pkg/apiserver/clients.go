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

type APIClient struct {
	// The base URL for the API server.
	BaseURL string
	// The HTTP client used to make requests.
	Client *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	// If baseURL is empty, set it to the default API server URL.
	if baseURL == "" {
		baseURL = APIServerUrl
	}

	// Create a new API client with the specified base URL and a default HTTP client.
	return &APIClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (c *APIClient) RegisterKubelet(kubelet *object.Kubelet) error {
	// Construct the URL for the Kubelet registration endpoint.
	url := c.BaseURL + KubeletRegisterURL

	// Convert the Kubelet object to JSON to be sent in the request body.
	kubeletJSON, err := json.Marshal(kubelet)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the Kubelet JSON as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(kubeletJSON))
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

func (c *APIClient) HeartbeatKubelet(kubelet *object.Kubelet) error {
	// Construct the URL for the Kubelet heartbeat endpoint.
	url := c.BaseURL + KubeletHeartbeatURL

	kubletBytes, err := json.Marshal(kubelet)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the node name as the body.
	req, err := http.NewRequest("POST", url, bytes.NewReader(kubletBytes))
	if err != nil {
		return err
	}
	// Set the content type to JSON.
	req.Header.Set("Content-Type", "application/json")

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

func (c *APIClient) GetNodes() ([]object.Kubelet, error) {
	var nodes []object.Kubelet
	err := c.getAndUnmarshalList(KubeletGetNodesURL, &nodes)

	return nodes, err
}

func (c *APIClient) CreatePod(pod *object.Pod) error {
	// Construct the URL for the Pod creation endpoint.
	url := c.BaseURL + PodCreateURL

	// Convert the Pod object to JSON to be sent in the request body.
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(podJSON))
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

func (c *APIClient) GetPods() ([]object.Pod, error) {
	var pods []object.Pod
	err := c.getAndUnmarshalList(PodGetURL, &pods)

	return pods, err
}

func (c *APIClient) AssignPodToNode(pod *object.Pod, nodeName string) error {
	// Construct the URL for the Pod assignment endpoint.
	url := c.BaseURL + PodAssignURL

	// Convert the Pod object to JSON to be sent in the request body.
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	// `nodeName` is passed as a query parameter.
	req, err := http.NewRequest(
		"POST",
		url+"?nodeName="+nodeName,
		bytes.NewBuffer(podJSON),
	)
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

func (c *APIClient) DeletePod(pod *object.Pod) error {
	// Construct the URL for the Pod deletion endpoint.
	url := c.BaseURL + PodDeleteURL

	// Convert the Pod object to JSON to be sent in the request body.
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(podJSON))
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

// TODO: CRUD on Service
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

func (c *APIClient) GetServices() ([]object.Service, error) {
	var svcs []object.Service
	err := c.getAndUnmarshalList(ServiceGetURL, &svcs)

	return svcs, err
}

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

func (c *APIClient) DeleteSingleDNS(dns *object.DNS) error {
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

func (c *APIClient) GetDNSResolve() ([]object.DNSResolveInfo, error) {
	// Get DNSResolveInfo from the API server.
	var dns []object.DNSResolveInfo
	err := c.getAndUnmarshalList(DNSGetResolveURL, &dns)

	// If the request fails, return an empty slice and the error.
	if err != nil {
		return []object.DNSResolveInfo{}, err
	}

	// If any of the domain does not end with ".", add it.
	for i := range dns {
		if dns[i].Host[len(dns[i].Host)-1] != '.' {
			dns[i].Host += "."
		}
	}

	return dns, nil
}
