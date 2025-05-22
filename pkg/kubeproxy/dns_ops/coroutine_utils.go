package dns_ops

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"slices"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/miekg/dns"
)

type DNSClient struct {
	// `resolveInfo` is a simple in-memory store after the DNS server starts.
	resolveInfo []object.DNSResolveInfo
	apiClient   *apiserver.APIClient
}

func NewDNSClient() (*DNSClient, error) {
	// Create a new DNS client.
	client := &DNSClient{
		resolveInfo: make([]object.DNSResolveInfo, 0),
		apiClient:   nil,
	}

	// Initialize the cluster services.
	if err := client.initializeClusterServices(); err != nil {
		log.Printf("Failed to initialize cluster services: %v", err)
		return nil, err
	}

	return client, nil
}

func (c *DNSClient) initializeClusterServices() error {
	// Initialize the api client.
	c.apiClient = apiserver.NewAPIClient("")

	// Get the list of all DNS records from the API server.
	resolveInfo, err := c.apiClient.GetDNSResolve()
	if err != nil {
		log.Printf("Failed to get DNS records: %v", err)
		panic(err)
	}

	// Assign the DNS records to the cluster services map.
	c.resolveInfo = resolveInfo

	return nil
}

func (c *DNSClient) HandleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Create a new DNS message for the response.
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true

	for _, q := range r.Question {
		// Convert the domain name to lowercase.
		domain := strings.ToLower(q.Name)
		log.Printf(
			"Received DNS query for domain: %s (%s)\n",
			domain,
			dns.TypeToString[q.Qtype],
		)

		// Only handle A records.
		if q.Qtype == dns.TypeA {
			// Use a bool variable to check if the domain is found.
			found := false

			// Check if the domain is in the resolveInfo array.
			for _, c := range c.resolveInfo {
				// Not found, continue to the next record.
				if c.Host != domain {
					continue
				}

				// Create a new A record
				rr, _ := dns.NewRR(fmt.Sprintf("%s 3600 IN A %s", domain, c.IP))
				m.Answer = append(m.Answer, rr)
				found = true

				break
			}

			// If the domain is not found in the resolveInfo array,
			// we will return an NXDOMAIN response.
			if !found {
				// Assign the NXDOMAIN response code.
				m.Rcode = dns.RcodeNameError

				// Optionally, log the domain not found.
				log.Printf("Domain not found: %s\n", domain)
			}
		}
	}

	// Sending the response back to the client.
	if err := w.WriteMsg(m); err != nil {
		log.Printf("failed to write DNS response: %v", err)
	}
}

func (c *DNSClient) MessageHandler(msg map[string]any) error {
	// Process the message sent on the queue.
	// Marshal the map to JSON bytes first
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message map: %v", err)
		return err
	}

	// Unmarshal the JSON bytes to the DNSResolveInfo object.
	var dnsResolveInfo object.DNSResolveInfo
	if err := json.Unmarshal(msgBytes, &dnsResolveInfo); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return err
	}

	// If the Host does not end with ".", add it.
	if dnsResolveInfo.Host[len(dnsResolveInfo.Host)-1] != '.' {
		dnsResolveInfo.Host += "."
	}

	log.Printf(
		"Received DNSResolveInfo: %s -> %s\n",
		dnsResolveInfo.Host,
		dnsResolveInfo.IP,
	)

	// Update the DNSResolveInfo in the resolveInfo array.
	// NOTE: if `IP` field is "0", it means to delete the record.
	if dnsResolveInfo.IP == "0" {
		// Delete the record from the resolveInfo array.
		for i, r := range c.resolveInfo {
			if r.Host == dnsResolveInfo.Host {
				// Remove the record from the array.
				c.resolveInfo = slices.Delete(c.resolveInfo, i, i+1)
				// Log the deletion.
				log.Printf("Deleted DNS record: %s\n", dnsResolveInfo.Host)

				break
			}
		}
	} else {
		// Use a bool variable to check if the record already exists.
		exists := false

		// Check if the record already exists.
		for i, r := range c.resolveInfo {
			if r.Host == dnsResolveInfo.Host {
				// Update the existing record.
				c.resolveInfo[i].IP = dnsResolveInfo.IP
				exists = true
				// Log the update.
				log.Printf("Updated DNS record: %s -> %s\n", dnsResolveInfo.Host, dnsResolveInfo.IP)

				break
			}
		}

		// If the record does not exist, add it to the resolveInfo array.
		if !exists {
			c.resolveInfo = append(c.resolveInfo, dnsResolveInfo)
			log.Printf("Added new DNS record: %s -> %s\n", dnsResolveInfo.Host, dnsResolveInfo.IP)
		}
	}

	return nil
}
