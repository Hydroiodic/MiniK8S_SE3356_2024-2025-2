package interfaces

import (
	"context"
	"os"

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
			Labels: map[string]string{
				object.SERVICE_INTERNEL_LABEL: "dns",
			},
		},
		Spec: object.ServiceSpec{
			Selector: map[string]string{},
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
					IP:   os.Getenv("APISERVER_URL"),
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
			Labels: map[string]string{
				object.SERVICE_INTERNEL_LABEL: "proxy",
			},
		},
		Spec: object.ServiceSpec{
			Selector: map[string]string{},
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
					IP:   os.Getenv("APISERVER_URL"),
					Port: 5301,
				},
			},
		},
	}
)

func GetInternalDNSService() *object.Service {
	internalDNSService.Status.Endpoints[0].IP = os.Getenv("APISERVER_URL")
	return internalDNSService
}

func GetInternalProxyService() *object.Service {
	internalProxyService.Status.Endpoints[0].IP = os.Getenv("APISERVER_URL")
	return internalProxyService
}

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
