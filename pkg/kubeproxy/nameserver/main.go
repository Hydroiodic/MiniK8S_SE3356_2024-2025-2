package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/dns_ops"
	"github.com/miekg/dns"
)

func syncForwardingInfo(
	client *apiserver.APIClient,
	dc *dns_ops.DNSClient,
	pc *dns_ops.DynamicProxy,
) error {
	// Fetch the forwarding info from the API server.
	forwardingInfo, err := client.GetForwardingInfo()
	if err != nil {
		return err
	}

	// Update the DNS resolve info in the client.
	dc.UpdateRules(forwardingInfo.DNSInfo)
	// Update the dynamic proxy rules.
	pc.UpdateRules(forwardingInfo.ProxyRules)

	// Log a message indicating the update was successful.
	log.Println("Forwarding info updated successfully")

	return nil
}

func syncRoutine(
	client *apiserver.APIClient,
	dc *dns_ops.DNSClient,
	pc *dns_ops.DynamicProxy,
	stopChan <-chan struct{},
) {
	ticker := time.NewTicker(5 * time.Second) // Sync every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := syncForwardingInfo(client, dc, pc); err != nil {
				log.Printf("Error syncing forwarding info: %v\n", err)
			}
		case <-stopChan:
			log.Println("Stopping sync routine")
			return
		}
	}
}

func startDNSServer(dc *dns_ops.DNSClient) {
	// Register the DNS handler for the DNS server.
	dns.HandleFunc(".", dc.HandleDNSRequest)

	// Run DNS server
	server := &dns.Server{
		Addr:    ":5300",
		Net:     "udp",
		UDPSize: 65535,
	}

	// Log a message indicating the server has started.
	log.Println("Starting DNS server on :5300")

	// Start DNS server and listen for incoming requests.
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start DNS server: %v\n", err)
	}

	// Ensure the server shuts down gracefully
	defer func() {
		if err := server.Shutdown(); err != nil {
			log.Printf("Error shutting down DNS server: %v\n", err)
		}
	}()
}

func startProxyServer(pc *dns_ops.DynamicProxy) {
	// Set up the HTTP server with the dynamic proxy.
	server := &http.Server{
		Addr: ":5301",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				pc.HandleCONNECT(w, r)
			} else {
				pc.ServeHTTP(w, r)
			}
		}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the HTTP proxy server.
	log.Println("Starting proxy server on :5301")
	log.Fatal(server.ListenAndServe())
}

func main() {
	// Create a new DNS client.
	dc := dns_ops.NewDNSClient()
	// Create a new dynamic proxy instance.
	pc := dns_ops.NewDynamicProxy()
	// Create a new API client to interact with the API server.
	client := apiserver.NewAPIClient("")

	// Run the sync routine in a separate goroutine.
	stopChan := make(chan struct{})
	go syncRoutine(client, dc, pc, stopChan)

	// Start the DNS server in a separate goroutine.
	go startDNSServer(dc)

	// Start the HTTP proxy server in a separate goroutine.
	go startProxyServer(pc)

	// Wait for a signal to stop the servers and sync routine.
	<-stopChan
	close(stopChan)
}
