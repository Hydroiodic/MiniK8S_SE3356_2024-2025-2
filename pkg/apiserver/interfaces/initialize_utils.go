package interfaces

import (
	"context"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

// NOTE: We are reusing the abstraction `Service` to redirect DNS and proxy requests.

var (
	internalDNSService = &object.Service{
		Kind: "Service",
		Type: object.SERVICE_TYPE_CLUSTERIP_STR,
		Metadata: object.Metadata{
			Name:      "dns-service",
			Namespace: "default",
			Labels:    map[string]string{"internal": "dns"},
		},
		Spec: object.ServiceSpec{
			Selector: map[string]string{"internal": "dns"},
			Ports: []object.ServicePort{
				{
					Name:       "http",
					Port:       53,
					TargetPort: 5300,
				},
			},
		},
		Status: object.ServiceStatus{
			ClusterIP: object.DNSClusterIP,
			Endpoints: []object.Endpoint{
				{
					IP:   "192.168.1.5", // Must be the node's IP address
					Port: 5300,
				},
			},
		},
	}
	internalProxyService = &object.Service{
		Kind: "Service",
		Type: object.SERVICE_TYPE_CLUSTERIP_STR,
		Metadata: object.Metadata{
			Name:      "proxy-service",
			Namespace: "default",
			Labels:    map[string]string{"internal": "proxy"},
		},
		Spec: object.ServiceSpec{
			Selector: map[string]string{"internal": "proxy"},
			Ports: []object.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: 5301,
				},
			},
		},
		Status: object.ServiceStatus{
			ClusterIP: object.ProxyClusterIP,
			Endpoints: []object.Endpoint{
				{
					IP:   "192.168.1.5",
					Port: 5301,
				},
			},
		},
	}
)

func InitializeInternalServices() {
	// Create a new ServiceStore instance.
	st, err := object.NewServiceStore([]string{})
	if err != nil {
		panic("Failed to create service store: " + err.Error())
	}

	// Create a context for the service store.
	ctx := context.Background()

	// Initialize the internal DNS service.
	if err := st.AddService(ctx, internalDNSService, true); err != nil {
		panic("Failed to initialize internal DNS service: " + err.Error())
	}

	// Initialize the internal proxy service.
	if err := st.AddService(ctx, internalProxyService, true); err != nil {
		panic("Failed to initialize internal proxy service: " + err.Error())
	}
}
