package object

import (
	"context"
	"encoding/json"
	"path"

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
func (s *DNSStore) key(namespace, name string) string {
	return path.Join(
		etcd.DNSPrefix,
		namespace,
		name,
	)
}

// AddDNS adds or updates a DNS object in etcd.
func (s *DNSStore) AddDNS(ctx context.Context, dns *DNS) error {
	// Marshal the DNS object into JSON.
	data, err := json.Marshal(dns)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(
		ctx,
		s.key(dns.Metadata.Namespace, dns.Metadata.Name),
		string(data),
	)
}

// GetDNS retrieves a DNS object from etcd.
func (s *DNSStore) GetDNS(
	ctx context.Context,
	namespace, name string,
) (*DNS, error) {
	// Check if the host is empty.
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
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
func (s *DNSStore) DeleteDNS(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
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
