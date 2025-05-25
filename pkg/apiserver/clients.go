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

func (c *APIClient) GetServiceByName(
	serviceName string,
) (object.Service, error) {
	services, err := c.GetServices()
	if err != nil {
		return object.Service{}, err
	}

	for _, service := range services {
		if service.Metadata.Name == serviceName {
			return service, err
		}
	}

	return object.Service{}, fmt.Errorf(
		"not service named" + serviceName + "found",
	)
}

func (c *APIClient) GetPodByName(
	podName string,
	namespace string,
) (object.Pod, error) {
	pods, err := c.GetPods()
	if err != nil {
		return object.Pod{}, err
	}

	for _, pod := range pods {
		if pod.Metadata.Name == podName && pod.Metadata.Namespace == namespace {
			return pod, err
		}
	}

	return object.Pod{}, fmt.Errorf("not pod named" + podName + "found")
}

func (c *APIClient) GeHpayName(
	hpaName string,
	namespace string,
) (object.HorizontalPodAutoscaler, error) {
	hpas, err := c.GetHpas()
	if err != nil {
		return object.HorizontalPodAutoscaler{}, err
	}

	for _, hpa := range hpas {
		if hpa.Metadata.Name == hpaName &&
			hpa.Metadata.Namespace == namespace {
			return hpa, err
		}
	}

	return object.HorizontalPodAutoscaler{}, fmt.Errorf(
		"not HPA named" + hpaName + "found",
	)
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
		"not HPA named" + replicasetName + "found",
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

func (c *APIClient) GetHpas() ([]object.HorizontalPodAutoscaler, error) {
	// Construct the URL for the Kubelet get nodes endpoint.
	url := c.BaseURL + HpaGetURL

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
	var hpas []object.HorizontalPodAutoscaler
	if err := json.NewDecoder(resp.Body).Decode(&hpas); err != nil {
		// If parsing fails, return an error.
		return nil, err
	}

	return hpas, nil
}

func (c *APIClient) DeleteDns(dns *object.DNS) error {
	// Construct the URL for the Pod deletion endpoint.
	url := c.BaseURL + DNSDeleteURL

	// Convert the Pod object to JSON to be sent in the request body.
	dnsJSON, err := json.Marshal(dns)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
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
func (c *APIClient) CreateHpa(hpa *object.HorizontalPodAutoscaler) error {
	// Construct the URL for the Pod creation endpoint.
	url := c.BaseURL + HpaCreateURL

	// Convert the Pod object to JSON to be sent in the request body.
	hpaJSON, err := json.Marshal(hpa)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(hpaJSON))
	if err != nil {
		return err
	}

	fmt.Println(hpa)

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

func (c *APIClient) CreateDns(dns *object.DNS) error {
	// Construct the URL for the Pod creation endpoint.
	url := c.BaseURL + DNSDeleteURL

	// Convert the Pod object to JSON to be sent in the request body.
	dnsJSON, err := json.Marshal(dns)
	if err != nil {
		return err
	}

	// Create a new HTTP POST request with the pod as the body.
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

func HasMatchingLabels(
	rsLabels, podLabels map[string]string,
) bool {
	for key, value := range rsLabels {
		if podValue, exists := podLabels[key]; exists && podValue == value {
			return true
		}
	}

	return false
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

func (c *APIClient) DeletePodByName(name string, namespace string) error {
	// Construct the URL for the Pod deletion endpoint.
	url := c.BaseURL + PodDeleteURL
	pod, err := c.GetPodByName(name, namespace)

	if err != nil {
		return err
	}
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
