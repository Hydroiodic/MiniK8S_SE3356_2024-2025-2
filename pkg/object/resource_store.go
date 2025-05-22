package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// ResourceStore is used to manage pod resource metrics in etcd.
type ResourceStore struct {
	etcdClient *etcd.Client
}

// NewResourceStore creates a new ResourceStore with the given etcd endpoints.
func NewResourceStore(etcdEndpoints []string) (*ResourceStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &ResourceStore{etcdClient: client}, nil
}

// Close is used to close the etcd client connection.
func (s *ResourceStore) Close() error {
	return s.etcdClient.Close()
}

// key returns the key for resource metrics in etcd using pod name and namespace.
func (s *ResourceStore) key(containerID string) string {
	return path.Join(etcd.ResourcePrefix, containerID)
}

// AddContainerResource is used to add pod resource metrics to etcd.
func (s *ResourceStore) AddContainerResource(
	ctx context.Context,
	containerID string,
	stat *ContainerStat,
) error {
	jsonData, err := json.Marshal(stat)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(ctx, s.key(containerID), string(jsonData))
}

// GetContainerResource is used to retrieve pod resource metrics from etcd.
func (s *ResourceStore) GetContainerResource(
	ctx context.Context,
	containerID string,
) (*ContainerStat, error) {
	data, err := s.etcdClient.Get(ctx, s.key(containerID))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var stat ContainerStat
	if err := json.Unmarshal([]byte(data), &stat); err != nil {
		return nil, err
	}

	return &stat, nil
}

// UpdateContainerResource is used to update pod resource metrics in etcd.
func (s *ResourceStore) UpdateContainerResource(
	ctx context.Context,
	containerID string,
	stat *ContainerStat,
) error {
	return s.AddContainerResource(ctx, containerID, stat)
}

// DeleteResourceMetrics is used to delete pod resource metrics from etcd.
func (s *ResourceStore) DeleteResourceMetrics(
	ctx context.Context,
	containerID string,
) error {
	return s.etcdClient.Delete(ctx, s.key(containerID))
}

// ListPodResources is used to list all pod resource metrics in etcd.
func (s *ResourceStore) ListPodResources(
	ctx context.Context,
	namespace, podName string,
) ([]*ContainerStat, error) {
	// Key for the pod in etcd.
	podKey := path.Join(etcd.PodPrefix, namespace, podName)
	// Get the pod from etcd.
	data, err := s.etcdClient.Get(ctx, podKey)
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	// Unmarshal the pod data.
	var pod Pod
	if err := json.Unmarshal([]byte(data), &pod); err != nil {
		return nil, err
	}

	// Get the container IDs from the pod.
	var containerIDs []string
	for _, container := range pod.Spec.Containers {
		containerIDs = append(containerIDs, container.ID)
	}

	// Use containerIDs to get the resource metrics.
	var statList []*ContainerStat

	for _, v := range containerIDs {
		// Get the resource metrics for each container.
		data, err := s.GetContainerResource(ctx, v)
		if data == nil || err != nil {
			continue
		}

		// Append the resource metrics to the list.
		statList = append(statList, data)
	}

	return statList, nil
}

// ListNodeResources is used to list all pod resource metrics in a specific node.
func (s *ResourceStore) ListNodeResources(
	ctx context.Context,
	nodeName string,
) ([]*PodResourceMetrics, error) {
	// TODO
	return nil, nil
}
