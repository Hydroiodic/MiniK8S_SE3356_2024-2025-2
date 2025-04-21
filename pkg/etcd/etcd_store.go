package etcd

import (
	"context"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	containerPrefix = "/minik8s/containers/"
)

type ContainerStore struct {
	etcdClient *EtcdClient
}

func NewContainerStore(etcdEndpoints []string) (*ContainerStore, error) {
	client, err := NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}
	return &ContainerStore{etcdClient: client}, nil
}

func (s *ContainerStore) Close() error {
	return s.etcdClient.Close()
}

func (s *ContainerStore) key(id string) string {
	return path.Join(containerPrefix, id)
}

// AddContainer 添加容器元数据
func (s *ContainerStore) AddContainer(ctx context.Context, container *object.Container) error {
	jsonData, err := container.ToJSON()
	if err != nil {
		return err
	}
	return s.etcdClient.Put(ctx, s.key(container.ID), jsonData)
}

// GetContainer 获取容器元数据
func (s *ContainerStore) GetContainer(ctx context.Context, id string) (*object.Container, error) {
	data, err := s.etcdClient.Get(ctx, s.key(id))
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}
	return object.ContainerFromJSON(data)
}

// UpdateContainer 更新容器元数据
func (s *ContainerStore) UpdateContainer(ctx context.Context, container *object.Container) error {
	return s.AddContainer(ctx, container)
}

// DeleteContainer 删除容器元数据
func (s *ContainerStore) DeleteContainer(ctx context.Context, id string) error {
	return s.etcdClient.Delete(ctx, s.key(id))
}

// ListContainers 列出所有容器元数据
func (s *ContainerStore) ListContainers(ctx context.Context) ([]*object.Container, error) {
	kvs, err := s.etcdClient.List(ctx, containerPrefix)
	if err != nil {
		return nil, err
	}

	var containers []*object.Container
	for _, v := range kvs {
		container, err := object.ContainerFromJSON(v)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}
	return containers, nil
}

// WatchContainers 监听容器元数据变化
func (s *ContainerStore) WatchContainers(ctx context.Context) (chan *object.Container, error) {
	watchChan := s.etcdClient.cli.Watch(ctx, containerPrefix, clientv3.WithPrefix())
	containerChan := make(chan *object.Container)

	go func() {
		defer close(containerChan)
		for resp := range watchChan {
			for _, ev := range resp.Events {
				if ev.Type == clientv3.EventTypePut {
					container, err := object.ContainerFromJSON(string(ev.Kv.Value))
					if err == nil {
						containerChan <- container
					}
				}
			}
		}
	}()

	return containerChan, nil
}
