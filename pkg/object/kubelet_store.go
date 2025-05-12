package object

import (
	"context"
	"encoding/json"
	"path"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// KubeletStore is used to manage kubelet metadata in etcd.
type KubeletStore struct {
	etcdClient *etcd.Client
}

// NewKubeletStore creates a new KubeletStore with the given etcd endpoints.
func NewKubeletStore(etcdEndpoints []string) (*KubeletStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &KubeletStore{etcdClient: client}, nil
}

// Close is used to close the etcd client connection.
func (s *KubeletStore) Close() error {
	return s.etcdClient.Close()
}

// key returns the key for kubelet in etcd using its name.
func (s *KubeletStore) key(name string) string {
	return path.Join(etcd.KubeletPrefix, name)
}

// AddKubelet is used to add kubelet metadata to etcd.
func (s *KubeletStore) AddKubelet(
	ctx context.Context,
	kubelet *Kubelet,
) error {
	kubelet.Mu.Lock()
	defer kubelet.Mu.Unlock()

	// Update timestamps
	kubelet.LastUpdateTime = time.Now()
	if kubelet.StartTime.IsZero() {
		kubelet.StartTime = time.Now()
	}

	jsonData, err := json.Marshal(kubelet)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(ctx, s.key(kubelet.Config.Name), string(jsonData))
}

// GetKubelet is used to retrieve kubelet metadata from etcd.
func (s *KubeletStore) GetKubelet(
	ctx context.Context,
	name string,
) (*Kubelet, error) {
	data, err := s.etcdClient.Get(ctx, s.key(name))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var kubelet Kubelet
	err = json.Unmarshal([]byte(data), &kubelet)
	if err != nil {
		return nil, err
	}

	return &kubelet, nil
}

// UpdateKubelet is used to update kubelet metadata in etcd.
func (s *KubeletStore) UpdateKubelet(
	ctx context.Context,
	kubelet *Kubelet,
) error {
	kubelet.Mu.Lock()
	defer kubelet.Mu.Unlock()

	// Update the last update time
	kubelet.LastUpdateTime = time.Now()

	// Calculate runtime duration
	kubelet.Runtime = time.Since(kubelet.StartTime)

	jsonData, err := json.Marshal(kubelet)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(ctx, s.key(kubelet.Config.Name), string(jsonData))
}

// DeleteKubelet is used to delete kubelet metadata from etcd.
func (s *KubeletStore) DeleteKubelet(ctx context.Context, name string) error {
	return s.etcdClient.Delete(ctx, s.key(name))
}

// ListKubelets is used to list all kubelet metadata in etcd.
func (s *KubeletStore) ListKubelets(
	ctx context.Context,
) ([]*Kubelet, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.KubeletPrefix)
	if err != nil {
		return nil, err
	}

	var kubelets []*Kubelet

	for _, v := range kvs {
		var kubelet Kubelet
		err := json.Unmarshal([]byte(v), &kubelet)
		if err != nil {
			return nil, err
		}

		kubelets = append(kubelets, &kubelet)
	}

	return kubelets, nil
}

// UpdateKubeletStatus updates the status of a kubelet in etcd.
func (s *KubeletStore) UpdateKubeletStatus(
	ctx context.Context,
	name, status string,
) error {
	kubelet, err := s.GetKubelet(ctx, name)
	if err != nil {
		return err
	}

	kubelet.Mu.Lock()
	defer kubelet.Mu.Unlock()

	kubelet.Status = status
	kubelet.LastUpdateTime = time.Now()
	kubelet.Runtime = time.Since(kubelet.StartTime)

	return s.UpdateKubelet(ctx, kubelet)
}

// UpdateKubeletPods updates the pods of a kubelet in etcd.
func (s *KubeletStore) UpdateKubeletPods(
	ctx context.Context,
	name string,
	pods []Pod,
) error {
	kubelet, err := s.GetKubelet(ctx, name)
	if err != nil {
		return err
	}

	kubelet.Mu.Lock()
	defer kubelet.Mu.Unlock()

	kubelet.Pods = pods
	kubelet.LastUpdateTime = time.Now()
	kubelet.Runtime = time.Since(kubelet.StartTime)

	return s.UpdateKubelet(ctx, kubelet)
}
