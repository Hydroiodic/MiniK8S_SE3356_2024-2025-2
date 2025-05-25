package object

import (
	"context"
	"encoding/json"
	"fmt"
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
	prefix := etcd.ValidServicePrefix
	if !ready {
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

func (s *ServiceStore) GetServiceWithoutStatus(
	ctx context.Context,
	namespace, name string,
) (*Service, error) {
	// Get the service data from Etcd.
	dataNotReady, err := s.etcdClient.Get(ctx, s.key(namespace, name, false))
	if err != nil {
		return nil, err
	}

	dataReady, err := s.etcdClient.Get(ctx, s.key(namespace, name, true))
	if err != nil {
		return nil, err
	}

	// Only one of the two data should be present.
	if dataNotReady == "" && dataReady == "" {
		return nil, fmt.Errorf(
			"service not found: %s/%s",
			namespace,
			name,
		)
	}

	if dataNotReady != "" && dataReady != "" {
		return nil, fmt.Errorf(
			"service found in both states: %s/%s",
			namespace,
			name,
		)
	}

	// Choose the correct data based on the ready status.
	var data string
	if dataNotReady != "" {
		data = dataNotReady
	} else {
		data = dataReady
	}

	// Unmarshal the data into a Service object.
	var svc Service
	if err := json.Unmarshal([]byte(data), &svc); err != nil {
		return nil, err
	}

	return &svc, nil
}

// GetService 从 Etcd 中获取对应的 Service 元数据
func (s *ServiceStore) GetService(
	ctx context.Context,
	namespace, name string,
	ready bool,
) (*Service, error) {
	// Get the service data from Etcd.
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name, ready))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	// Unmarshal the data into a Service object.
	var svc Service
	if err := json.Unmarshal([]byte(data), &svc); err != nil {
		return nil, err
	}

	return &svc, nil
}

func (s *ServiceStore) UpdateServiceWithoutStatus(
	ctx context.Context,
	svc *Service,
) error {
	// Get the service data from Etcd.
	dataNotReady, err := s.etcdClient.Get(
		ctx,
		s.key(svc.Metadata.Namespace, svc.Metadata.Name, false),
	)
	if err != nil {
		return err
	}

	dataReady, err := s.etcdClient.Get(
		ctx,
		s.key(svc.Metadata.Namespace, svc.Metadata.Name, true),
	)
	if err != nil {
		return err
	}

	// Only one of the two data should be present.
	if dataNotReady == "" && dataReady == "" {
		return fmt.Errorf(
			"service not found: %s/%s",
			svc.Metadata.Namespace,
			svc.Metadata.Name,
		)
	}

	if dataNotReady != "" && dataReady != "" {
		return fmt.Errorf(
			"service found in both states: %s/%s",
			svc.Metadata.Namespace,
			svc.Metadata.Name,
		)
	}

	nowReady := svc.CheckReady()
	prevReady := dataReady != ""

	// If the ready status does not changd, just update the service data.
	if nowReady == prevReady {
		return s.UpdateService(ctx, svc, nowReady)
	}

	// The status has changed, so we need to delete the old data and add the new one.
	// TODO: make the two database operations atomic.
	err = s.etcdClient.Delete(
		ctx,
		s.key(svc.Metadata.Namespace, svc.Metadata.Name, prevReady),
	)
	if err != nil {
		return err
	}

	// Add the new service data.
	err = s.AddService(ctx, svc, nowReady)
	if err != nil {
		return err
	}

	return nil
}

// UpdateService 更新 Etcd 中的 Service 元数据（内部复用 AddService）
func (s *ServiceStore) UpdateService(
	ctx context.Context,
	svc *Service,
	ready bool,
) error {
	return s.AddService(ctx, svc, ready)
}

func (s *ServiceStore) DeleteServiceWithoutStatus(
	ctx context.Context,
	namespace, name string,
) error {
	// Get the service data from Etcd.
	dataNotReady, err := s.etcdClient.Get(ctx, s.key(namespace, name, false))
	if err != nil {
		return err
	}

	dataReady, err := s.etcdClient.Get(ctx, s.key(namespace, name, true))
	if err != nil {
		return err
	}

	// Only one of the two data should be present.
	if dataNotReady == "" && dataReady == "" {
		return fmt.Errorf(
			"service not found: %s/%s",
			namespace,
			name,
		)
	}

	if dataNotReady != "" && dataReady != "" {
		return fmt.Errorf(
			"service found in both states: %s/%s",
			namespace,
			name,
		)
	}

	// Delete the service data from Etcd.
	err = s.etcdClient.Delete(ctx, s.key(namespace, name, dataReady != ""))
	if err != nil {
		return err
	}

	return nil
}

// DeleteService 从 Etcd 中删除指定的 Service 元数据
func (s *ServiceStore) DeleteService(
	ctx context.Context,
	namespace, name string,
	ready bool,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name, ready))
}

func (s *ServiceStore) ListServicesWithoutStatus(
	ctx context.Context,
) ([]*Service, error) {
	// Get all services from Etcd.
	kvs, err := s.etcdClient.List(ctx, etcd.AllServicePrefix)
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

	// Get all services from Etcd.
	kvs, err := s.etcdClient.List(ctx, servicePrefix)
	if err != nil {
		return nil, err
	}

	var services []*Service
	// Iterate over the key-value pairs and unmarshal them into Service objects.
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
	// Iterate over the key-value pairs and unmarshal them into Service objects.
	for _, v := range kvs {
		var svc Service
		if err := json.Unmarshal([]byte(v), &svc); err != nil {
			return nil, err
		}

		services = append(services, &svc)
	}

	return services, nil
}
