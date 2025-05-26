package main

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func main() {
	// Create a new API client.
	client := apiserver.NewAPIClient("")

	// Create a example DNS object.
	dns := object.DNS{
		Kind: "DNS",
		Metadata: object.Metadata{
			Name:      "example-dns",
			Namespace: "default",
			Labels:    map[string]string{"app": "example"},
		},
		Spec: object.DNSSpec{
			Host: "example.com",
			Paths: []object.DNSPath{
				{
					Path:        "/",
					ServiceName: "example-service",
					ServiceIP:   "127.0.0.1",
					ServicePort: 80,
				},
			},
		},
	}

	// Add the DNS object to the API server.
	if err := client.AddDNS(&dns); err != nil {
		panic(err)
	}
}
