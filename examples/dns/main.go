package main

import (
	"log"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/dns_ops"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/miekg/dns"
)

func main() {
	// Create a new DNS client.
	client, err := dns_ops.NewDNSClient()
	if err != nil {
		panic(err)
	}

	// Create a go coroutine to handle Message Queue.
	go func() {
		if err := mqtemplate.ConsumeMessageOnQueue(
			mqtemplate.UpdateHostQueueName,
			client.MessageHandler,
		); err != nil {
			log.Printf("Error consuming message on queue: %v\n", err)
			panic(err)
		}
	}()

	// Register the DNS handler for the DNS server.
	dns.HandleFunc(".", client.HandleDNSRequest)

	// Run DNS server
	server := &dns.Server{
		Addr:    ":5333",
		Net:     "udp",
		UDPSize: 65535,
	}

	// Start the DNS server and listen for incoming requests.
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start DNS server: %v\n", err)
	}

	log.Println("DNS server started on port 53")

	// Ensure the server shuts down gracefully
	defer func() {
		if err := server.Shutdown(); err != nil {
			log.Printf("Error shutting down DNS server: %v\n", err)
		}
	}()
}
