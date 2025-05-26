package object

import (
	"context"
	"encoding/json"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// `ClusterIPStore` provides methods to manage ClusterIP objects in Etcd.
type ClusterIPStore struct {
	etcdClient *etcd.Client
}

// `NewClusterIPStore` creates a new `ClusterIPStore` instance.
func NewClusterIPStore(etcdEndpoints []string) (*ClusterIPStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &ClusterIPStore{etcdClient: client}, nil
}

// `Close` will close the Etcd client connection.
func (s *ClusterIPStore) Close() error {
	return s.etcdClient.Close()
}

func (s *ClusterIPStore) GetClusterIPStruct(
	ctx context.Context,
) (*ClusterIP, error) {
	data, err := s.etcdClient.Get(ctx, etcd.ClusterIPPrefix)
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var clusterIP ClusterIP
	if err := json.Unmarshal([]byte(data), &clusterIP); err != nil {
		return nil, err
	}

	return &clusterIP, nil
}

// ListClusterIPs will retrieve all ClusterIPs from Etcd.
func (s *ClusterIPStore) ListClusterIPs(
	ctx context.Context,
) ([]string, error) {
	// Get the ClusterIP object from Etcd.
	clusterIP, err := s.GetClusterIPStruct(ctx)
	if err != nil {
		return nil, err
	} else if clusterIP == nil {
		return make([]string, 0), nil
	}

	// Extract the ClusterIPs from the ClusterIP object.
	result := make([]string, 0, len(clusterIP.ClusterIPMap))
	for cip := range clusterIP.ClusterIPMap {
		result = append(result, cip)
	}

	return result, nil
}

// `SetClusterIP` will store a ClusterIP in Etcd.
func (s *ClusterIPStore) SetClusterIP(
	ctx context.Context,
	clusterIP string,
) error {
	// First, we get the existing ClusterIP object.
	existingClusterIP, err := s.GetClusterIPStruct(ctx)
	if err != nil {
		return err
	}

	// If it doesn't exist, create a new one.
	if existingClusterIP == nil {
		existingClusterIP = &ClusterIP{
			ClusterIPMap: make(map[string]any),
		}
	}

	// Add or update the ClusterIP in the map.
	existingClusterIP.ClusterIPMap[clusterIP] = nil

	// Marshal the ClusterIP object to JSON.
	data, err := json.Marshal(existingClusterIP)
	if err != nil {
		return err
	}

	// Store the ClusterIP object in Etcd.
	if err := s.etcdClient.Put(ctx, etcd.ClusterIPPrefix, string(data)); err != nil {
		return err
	}

	return nil
}

// `DeleteClusterIP` will delete a ClusterIP from Etcd.
func (s *ClusterIPStore) DeleteClusterIP(
	ctx context.Context,
	cip string,
) error {
	// First, we get the existing ClusterIP object.
	existingClusterIP, err := s.GetClusterIPStruct(ctx)
	if err != nil {
		return err
	}

	// If it doesn't exist, return nil.
	if existingClusterIP == nil {
		return nil
	}

	// Remove the ClusterIP from the map.
	delete(existingClusterIP.ClusterIPMap, cip)

	// Serialize the updated ClusterIP object to JSON.
	data, err := json.Marshal(existingClusterIP)
	if err != nil {
		return err
	}

	// Store the updated ClusterIP object in Etcd.
	if err := s.etcdClient.Put(ctx, etcd.ClusterIPPrefix, string(data)); err != nil {
		return err
	}

	return nil
}
