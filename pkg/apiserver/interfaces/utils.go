package interfaces

import (
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func retrievePodsName(pods []object.Pod) []string {
	// Create a slice to store the pod names.
	podNames := make([]string, 0)

	// Iterate over the pods and append their names to the slice.
	for _, pod := range pods {
		podNames = append(
			podNames,
			path.Join(pod.Metadata.Namespace, pod.Metadata.Name),
		)
	}

	return podNames
}

func retrieveServicesName(services []object.Service) []string {
	// Create a slice to store the service names.
	serviceNames := make([]string, 0)

	// Iterate over the services and append their names to the slice.
	for _, service := range services {
		serviceNames = append(
			serviceNames,
			path.Join(service.Metadata.Namespace, service.Metadata.Name),
		)
	}

	return serviceNames
}

func newSimpleDNS(domain, host, name string) object.DNS { //nolint
	// Create a simple DNS object with the given domain and host.
	// TODO: Refine the DNS object later.
	return object.DNS{
		Kind: "DNS",
		Metadata: object.Metadata{
			Name:      name,
			Namespace: "default",
			Labels:    nil,
		},
		Spec: object.DNSSpec{
			Host: domain,
			Paths: []object.DNSPath{
				{
					Path:        "/",
					ServiceName: name,
					ServiceIP:   host,
					ServicePort: 80,
				},
			},
		},
	}
}
