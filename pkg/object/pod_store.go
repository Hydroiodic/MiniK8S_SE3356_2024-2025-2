package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// PodStore is used to manage pod metadata in etcd.
type PodStore struct {
	etcdClient *etcd.Client
}

// NewPodStore creates a new PodStore with the given etcd endpoints.
func NewPodStore(etcdEndpoints []string) (*PodStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &PodStore{etcdClient: client}, nil
}

// Close is used to close the etcd client connection.
func (s *PodStore) Close() error {
	return s.etcdClient.Close()
}

// key returns the key for pod in etcd using its name and namespace.
func (s *PodStore) key(namespace, name string) string {
	return path.Join(etcd.PodPrefix, namespace, name)
}

// AddPod is used to add pod metadata to etcd.
func (s *PodStore) AddPod(
	ctx context.Context,
	pod *Pod,
) error {
	jsonData, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(
		ctx,
		s.key(pod.Metadata.Namespace, pod.Metadata.Name),
		string(jsonData),
	)
}

// GetPod is used to retrieve pod metadata from etcd.
func (s *PodStore) GetPod(
	ctx context.Context,
	namespace, name string,
) (*Pod, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var pod Pod

	err = json.Unmarshal([]byte(data), &pod)
	if err != nil {
		return nil, err
	}

	return &pod, nil
}

// UpdatePod is used to update pod metadata in etcd.
func (s *PodStore) UpdatePod(
	ctx context.Context,
	pod *Pod,
) error {
	return s.AddPod(ctx, pod)
}

// DeletePod is used to delete pod metadata from etcd.
func (s *PodStore) DeletePod(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
}

// ListPods is used to list all pod metadata in etcd.
func (s *PodStore) ListPods(
	ctx context.Context,
) ([]*Pod, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.PodPrefix)
	if err != nil {
		return nil, err
	}

	var pods []*Pod

	for _, v := range kvs {
		var pod Pod

		err := json.Unmarshal([]byte(v), &pod)
		if err != nil {
			return nil, err
		}

		pods = append(pods, &pod)
	}

	return pods, nil
}

// ListPodsInNamespace is used to list all pod metadata in a specific namespace.
func (s *PodStore) ListPodsInNamespace(
	ctx context.Context,
	namespace string,
) ([]*Pod, error) {
	kvs, err := s.etcdClient.List(ctx, path.Join(etcd.PodPrefix, namespace))
	if err != nil {
		return nil, err
	}

	var pods []*Pod

	for _, v := range kvs {
		var pod Pod
		err := json.Unmarshal([]byte(v), &pod)

		if err != nil {
			return nil, err
		}

		pods = append(pods, &pod)
	}

	return pods, nil
}
