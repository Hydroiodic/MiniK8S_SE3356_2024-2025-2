package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// ServicePrefix 是存储 Service 对象的前缀，和 PodPrefix 类似
const ServicePrefix = "/miniK8s/services"

// ServiceStore 提供对 Service 元数据的增删改查
type ServiceStore struct {
	etcdClient *etcd.Client
}

// NewServiceStore 使用给定的 etcdEndpoints 创建一个 ServiceStore
func NewServiceStore(etcdEndpoints []string) (*ServiceStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &ServiceStore{etcdClient: client}, nil
}

// Close 用于关闭 Etcd 客户端连接
func (s *ServiceStore) Close() error {
	return s.etcdClient.Close()
}

// key 根据 Service 的 Namespace 和 Name 生成在 Etcd 中的存储路径
func (s *ServiceStore) key(namespace, name string) string {
	return path.Join(ServicePrefix, namespace, name)
}

// AddService 将 Service 元数据存储到 Etcd
func (s *ServiceStore) AddService(
	ctx context.Context,
	svc *Service,
) error {
	jsonData, err := json.Marshal(svc)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(
		ctx,
		s.key(svc.Metadata.Namespace, svc.Metadata.Name),
		string(jsonData),
	)
}

// GetService 从 Etcd 中获取对应的 Service 元数据
func (s *ServiceStore) GetService(
	ctx context.Context,
	namespace, name string,
) (*Service, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
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
) error {
	return s.AddService(ctx, svc)
}

// DeleteService 从 Etcd 中删除指定的 Service 元数据
func (s *ServiceStore) DeleteService(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
}

// ListServices 列出所有 Service 元数据
func (s *ServiceStore) ListServices(
	ctx context.Context,
) ([]*Service, error) {
	kvs, err := s.etcdClient.List(ctx, ServicePrefix)
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
) ([]*Service, error) {
	kvs, err := s.etcdClient.List(ctx, path.Join(ServicePrefix, namespace))
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
