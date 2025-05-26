package object

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// DNSStore manages DNS objects in etcd.
type DNSStore struct {
	etcdClient *etcd.Client
}

// NewDNSStore creates a new DNSStore with the given etcd endpoints.
func NewDNSStore(etcdEndpoints []string) (*DNSStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &DNSStore{etcdClient: client}, nil
}

// Close closes the etcd client connection.
func (s *DNSStore) Close() error {
	return s.etcdClient.Close()
}

// key returns the key for DNS in etcd using its host.
func (s *DNSStore) key(host string) string {
	// Trim the dot at the end of the host if it exists.
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	return path.Join(etcd.DNSPrefix, host)
}

// AddDNS adds or updates a DNS object in etcd.
func (s *DNSStore) AddDNS(ctx context.Context, dns *DNS) error {
	// Marshal the DNS object into JSON.
	data, err := json.Marshal(dns)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(ctx, s.key(dns.Spec.Host), string(data))
}

func (s *DNSStore) CombineDNS(ctx context.Context, dns *DNS) error {
	// Get the DNS object.
	dnsOld, err := s.GetDNS(ctx, dns.Spec.Host)
	if err != nil {
		return err
	}

	// Check if the DNS object is nil.
	if dnsOld == nil {
		// If the DNS object is nil, create a new one.
		return s.AddDNS(ctx, dns)
	}

	// Add the new path to the DNS object.
	for _, newPath := range dns.Spec.Paths {
		// `exists` is a flag to check if the path already exists.
		exists := false

		// Check if the path already exists.
		for _, oldPath := range dnsOld.Spec.Paths {
			if newPath.Path == oldPath.Path {
				exists = true
				break
			}
		}

		// If the path exists, skip it and report a warning.
		if exists {
			fmt.Printf(
				"Path %s already exists in DNS object for host %s\n",
				newPath.Path,
				dns.Spec.Host,
			)

			continue
		}

		// Append the new path to the existing paths.
		dnsOld.Spec.Paths = append(dnsOld.Spec.Paths, newPath)
	}

	// Update the DNS object in etcd.
	return s.AddDNS(ctx, dnsOld)
}

// GetDNS retrieves a DNS object from etcd.
func (s *DNSStore) GetDNS(ctx context.Context, host string) (*DNS, error) {
	// Check if the host is empty.
	data, err := s.etcdClient.Get(ctx, s.key(host))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	// Unmarshal the JSON data into a DNS object.
	var dns DNS
	if err = json.Unmarshal([]byte(data), &dns); err != nil {
		return nil, err
	}

	return &dns, nil
}

// DeleteDNS removes a DNS object from etcd.
func (s *DNSStore) DeleteDNS(ctx context.Context, host string) error {
	return s.etcdClient.Delete(ctx, s.key(host))
}

func (s *DNSStore) DeleteSinglePathDNS(
	ctx context.Context,
	host string,
	path string,
) error {
	// Get the DNS object.
	dns, err := s.GetDNS(ctx, host)
	if err != nil {
		return err
	}

	// Check if the DNS object is nil.
	if dns == nil {
		return fmt.Errorf("DNS object not found for host: %s", host)
	}

	// Remove the specified path from the DNS object.
	for i, p := range dns.Spec.Paths {
		if p.Path == path {
			dns.Spec.Paths = slices.Delete(dns.Spec.Paths, i, i+1)
			break
		}
	}

	// If the length of paths is 0, delete the DNS object.
	if len(dns.Spec.Paths) == 0 {
		return s.DeleteDNS(ctx, host)
	}

	// Update the DNS object in etcd.
	return s.AddDNS(ctx, dns)
}

// ListDNS lists all DNS objects in etcd.
func (s *DNSStore) ListDNS(ctx context.Context) ([]*DNS, error) {
	// List all keys in the DNS prefix.
	kvs, err := s.etcdClient.List(ctx, etcd.DNSPrefix)
	if err != nil {
		return nil, err
	}

	// Use a slice to store the DNS objects.
	var dnsList []*DNS

	// Iterate over the key-value pairs and unmarshal each one into a DNS object.
	for _, v := range kvs {
		// Unmarshal the JSON data into a DNS object.
		var d DNS
		if err := json.Unmarshal([]byte(v), &d); err != nil {
			return nil, err
		}

		// Append the DNS object to the list.
		dnsList = append(dnsList, &d)
	}

	return dnsList, nil
}
