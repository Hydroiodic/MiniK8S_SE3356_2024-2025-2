package object

import (
	"context"
	"encoding/json"
	"path"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// KubeProxyStore is used to manage kubeproxy metadata in etcd.
type KubeProxyStore struct {
	etcdClient *etcd.Client
}

// NewKubeProxyStore creates a new KubeProxyStore with the given etcd endpoints.
func NewKubeProxyStore(etcdEndpoints []string) (*KubeProxyStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &KubeProxyStore{etcdClient: client}, nil
}

// Close closes the etcd client connection.
func (s *KubeProxyStore) Close() error {
	return s.etcdClient.Close()
}

// key returns the key for kubeproxy in etcd using its name.
func (s *KubeProxyStore) key(name string) string {
	return path.Join(etcd.KubeProxyPrefix, name)
}

// AddKubeProxy adds kubeproxy metadata to etcd.
func (s *KubeProxyStore) AddKubeProxy(
	ctx context.Context,
	kubeproxy *KubeProxy,
) error {
	kubeproxy.Mu.Lock()
	defer kubeproxy.Mu.Unlock()

	kubeproxy.LastUpdateTime = time.Now()

	jsonData, err := json.Marshal(kubeproxy)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(ctx, s.key(kubeproxy.Config.Name), string(jsonData))
}

// GetKubeProxy retrieves kubeproxy metadata from etcd.
func (s *KubeProxyStore) GetKubeProxy(
	ctx context.Context,
	name string,
) (*KubeProxy, error) {
	data, err := s.etcdClient.Get(ctx, s.key(name))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var kubeproxy KubeProxy

	err = json.Unmarshal([]byte(data), &kubeproxy)
	if err != nil {
		return nil, err
	}

	return &kubeproxy, nil
}

// UpdateKubeProxy updates kubeproxy metadata in etcd.
func (s *KubeProxyStore) UpdateKubeProxy(
	ctx context.Context,
	kubeproxy *KubeProxy,
) error {
	kubeproxy.Mu.Lock()
	defer kubeproxy.Mu.Unlock()

	kubeproxy.LastUpdateTime = time.Now()

	jsonData, err := json.Marshal(kubeproxy)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(ctx, s.key(kubeproxy.Config.Name), string(jsonData))
}

// DeleteKubeProxy deletes kubeproxy metadata from etcd.
func (s *KubeProxyStore) DeleteKubeProxy(
	ctx context.Context,
	name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(name))
}

// ListKubeProxies lists all kubeproxy metadata in etcd.
func (s *KubeProxyStore) ListKubeProxies(
	ctx context.Context,
) ([]*KubeProxy, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.KubeProxyPrefix)
	if err != nil {
		return nil, err
	}

	var kubeproxies []*KubeProxy

	for _, v := range kvs {
		var kubeproxy KubeProxy

		err := json.Unmarshal([]byte(v), &kubeproxy)
		if err != nil {
			return nil, err
		}

		kubeproxies = append(kubeproxies, &kubeproxy)
	}

	return kubeproxies, nil
}

// UpdateKubeProxyServices updates the services of a kubeproxy in etcd.
func (s *KubeProxyStore) UpdateKubeProxyServices(
	ctx context.Context,
	name string,
	services []Service,
) error {
	kubeproxy, err := s.GetKubeProxy(ctx, name)
	if err != nil {
		return err
	}

	if kubeproxy == nil {
		return nil
	}

	kubeproxy.Mu.Lock()
	defer kubeproxy.Mu.Unlock()

	kubeproxy.Services = services
	kubeproxy.LastUpdateTime = time.Now()

	return s.UpdateKubeProxy(ctx, kubeproxy)
}

// UpdateKubeProxyHeartbeat updates the heartbeat (last update time) of a kubeproxy in etcd.
func (s *KubeProxyStore) UpdateKubeProxyHeartbeat(
	ctx context.Context,
	name string,
) error {
	kubeproxy, err := s.GetKubeProxy(ctx, name)
	if err != nil {
		return err
	}

	if kubeproxy == nil {
		return nil
	}

	kubeproxy.Mu.Lock()
	defer kubeproxy.Mu.Unlock()

	kubeproxy.LastUpdateTime = time.Now()

	return s.UpdateKubeProxy(ctx, kubeproxy)
}
