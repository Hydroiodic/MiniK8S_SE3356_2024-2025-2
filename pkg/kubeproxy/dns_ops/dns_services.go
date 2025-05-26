package dns_ops

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/miekg/dns"
)

type DNSClient struct {
	// `resolveInfo` is a simple in-memory store after the DNS server starts.
	rules []object.DNSResolveInfo
	mu    sync.RWMutex
}

func NewDNSClient() *DNSClient {
	// Create a new DNS client.
	return &DNSClient{
		rules: make([]object.DNSResolveInfo, 0),
		mu:    sync.RWMutex{},
	}
}

func (c *DNSClient) UpdateRules(newRules []object.DNSResolveInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update the rules with the new rules.
	c.rules = newRules

	// If the domain does not end with a dot, we will add a dot to the end.
	for i := range c.rules {
		c.rules[i].Host = strings.ToLower(c.rules[i].Host)
		if !strings.HasSuffix(c.rules[i].Host, ".") {
			c.rules[i].Host = c.rules[i].Host + "."
		}
	}

	// Log the updated rules.
	log.Printf("DNS rules updated: %d rules loaded\n", len(c.rules))

	for _, rule := range c.rules {
		log.Printf("Rule: %s -> %s\n", rule.Host, rule.IP)
	}
}

func (c *DNSClient) HandleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Create a new DNS message for the response.
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true

	// Make a copy of `rules` to avoid holding the lock for too long.
	c.mu.RLock()
	rulesCopy := make([]object.DNSResolveInfo, len(c.rules))
	copy(rulesCopy, c.rules)
	c.mu.RUnlock()

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

			// Check if the domain is in the `rules` array.
			for _, c := range rulesCopy {
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
			// we will return a SERVFAIL response so the client tries the next DNS.
			if !found {
				// Assign the SERVFAIL response code.
				m.Rcode = dns.RcodeServerFailure

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
