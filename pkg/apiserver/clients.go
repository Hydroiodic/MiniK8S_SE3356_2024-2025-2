package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func (c *APIClient) GetNodes() ([]object.Kubelet, error) {
	// Construct the URL for the Kubelet get nodes endpoint.
	url := c.BaseURL + KubeletGetNodesURL

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

	// 先完整读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 打印日志
	fmt.Println("Received bytes: ", string(bodyBytes))

	// 先解析外层字符串
	var raw string
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal as string: %v", err)
	}

	// 再解析内部的 JSON 数组
	var nodes []object.Kubelet
	if err := json.Unmarshal([]byte(raw), &nodes); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal as []object.Kubelet: %v",
			err,
		)
	}

	// Return the list of Kubelet objects.
	return nodes, nil
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

func (c *APIClient) GetPods() ([]object.Pod, error) {
	// Construct the URL for the Pod get endpoint.
	url := c.BaseURL + PodGetURL

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

	// Parse the response body as a list of Pod objects.
	var pods []object.Pod
	if err := json.NewDecoder(resp.Body).Decode(&pods); err != nil {
		// If parsing fails, return an error.
		return nil, err
	}

	// Return the list of Pod objects.
	return pods, nil
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

func (c *APIClient) DeletePodFromEtcd(pod *object.Pod) error {
	// Construct the URL for the Pod deletion endpoint.
	url := c.BaseURL + KubeletDeletePodURL

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
