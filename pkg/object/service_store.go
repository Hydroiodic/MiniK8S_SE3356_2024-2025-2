package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// ServiceStore provides methods to manage Service objects in Etcd.
type ServiceStore struct {
	etcdClient *etcd.Client
}

func NewServiceStore(etcdEndpoints []string) (*ServiceStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &ServiceStore{etcdClient: client}, nil
}

func (s *ServiceStore) Close() error {
	return s.etcdClient.Close()
}

// key 根据 Service 的 Namespace 和 Name 生成在 Etcd 中的存储路径
func (s *ServiceStore) key(namespace, name string, ready bool) string {
	var prefix string
	if ready {
		prefix = etcd.ValidServicePrefix
	} else {
		prefix = etcd.PendingServicePrefix
	}

	return path.Join(prefix, namespace, name)
}

// AddService 将 Service 元数据存储到 Etcd
func (s *ServiceStore) AddService(
	ctx context.Context,
	svc *Service,
	ready bool,
) error {
	jsonData, err := json.Marshal(svc)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(
		ctx,
		s.key(svc.Metadata.Namespace, svc.Metadata.Name, ready),
		string(jsonData),
	)
}

// GetService 从 Etcd 中获取对应的 Service 元数据
func (s *ServiceStore) GetService(
	ctx context.Context,
	namespace, name string,
	ready bool,
) (*Service, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name, ready))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var svc Service
	if err := json.Unmarshal([]byte(data), &svc); err != nil {
		return nil, err
	}

	return &svc, nil
}

// UpdateService 更新 Etcd 中的 Service 元数据（内部复用 AddService）
func (s *ServiceStore) UpdateService(
	ctx context.Context,
	svc *Service,
	ready bool,
) error {
	return s.AddService(ctx, svc, ready)
}

// DeleteService 从 Etcd 中删除指定的 Service 元数据
func (s *ServiceStore) DeleteService(
	ctx context.Context,
	namespace, name string,
	ready bool,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name, ready))
}

// ListServices 列出所有 Service 元数据
func (s *ServiceStore) ListServices(
	ctx context.Context,
	ready bool,
) ([]*Service, error) {
	// Get prefix based on the ready status.
	servicePrefix := etcd.ValidServicePrefix
	if !ready {
		servicePrefix = etcd.PendingServicePrefix
	}

	kvs, err := s.etcdClient.List(ctx, servicePrefix)
	if err != nil {
		return nil, err
	}

	var services []*Service

	for _, v := range kvs {
		var svc Service
		if err := json.Unmarshal([]byte(v), &svc); err != nil {
			return nil, err
		}

		services = append(services, &svc)
	}

	return services, nil
}

// ListServicesInNamespace 列出指定 Namespace 下的 Service 元数据
func (s *ServiceStore) ListServicesInNamespace(
	ctx context.Context,
	namespace string,
	ready bool,
) ([]*Service, error) {
	// Get prefix based on the ready status.
	servicePrefix := etcd.ValidServicePrefix
	if !ready {
		servicePrefix = etcd.PendingServicePrefix
	}

	// Use the namespace to filter the services.
	kvs, err := s.etcdClient.List(ctx, path.Join(servicePrefix, namespace))
	if err != nil {
		return nil, err
	}

	var services []*Service

	for _, v := range kvs {
		var svc Service
		if err := json.Unmarshal([]byte(v), &svc); err != nil {
			return nil, err
		}

		services = append(services, &svc)
	}

	return services, nil
}
