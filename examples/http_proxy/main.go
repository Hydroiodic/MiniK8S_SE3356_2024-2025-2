package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/dns_ops"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func main() {
	// Create a new dynamic proxy instance.
	proxy := dns_ops.NewDynamicProxy()

	// Initialize with some rules.
	initialRules := []object.ProxyRule{
		{
			Domain:     "example.com",
			PathPrefix: "/1",
			Target:     "https://10.0.0.1:443",
		},
		{
			Domain:     "localhost",
			PathPrefix: "/2",
			Target:     "http://www.baidu.com:80",
		},
	}
	proxy.UpdateRules(initialRules)

	// Set up the HTTP server with the dynamic proxy.
	server := &http.Server{
		Addr: ":5301",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				proxy.HandleCONNECT(w, r)
			} else {
				proxy.ServeHTTP(w, r)
			}
		}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the HTTP server.
	log.Println("Starting proxy server on :5301")
	log.Fatal(server.ListenAndServe())
}
