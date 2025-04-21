package etcd

import (
	"context"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ContainerStore is used to manage container metadata in etcd.
type ContainerStore struct {
	etcdClient *Client
}

// NewContainerStore creates a new ContainerStore with the given etcd endpoints.
func NewContainerStore(etcdEndpoints []string) (*ContainerStore, error) {
	client, err := NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &ContainerStore{etcdClient: client}, nil
}

// Close is used to close the etcd client connection.
func (s *ContainerStore) Close() error {
	return s.etcdClient.Close()
}

// keyPrefix returns the prefix for container keys in etcd.
func (*ContainerStore) key(id string) string {
	return path.Join(containerPrefix, id)
}

// AddContainer is used to add container metadata to etcd.
func (s *ContainerStore) AddContainer(
	ctx context.Context,
	container *object.Container,
) error {
	jsonData, err := container.ToJSON()
	if err != nil {
		return err
	}

	return s.etcdClient.Put(ctx, s.key(container.ID), jsonData)
}

// GetContainer is used to retrieve container metadata from etcd.
func (s *ContainerStore) GetContainer(
	ctx context.Context,
	id string,
) (*object.Container, error) {
	data, err := s.etcdClient.Get(ctx, s.key(id))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	return object.ContainerFromJSON(data)
}

// UpdateContainer is used to update container metadata in etcd.
func (s *ContainerStore) UpdateContainer(
	ctx context.Context,
	container *object.Container,
) error {
	return s.AddContainer(ctx, container)
}

// DeleteContainer is used to delete container metadata from etcd.
func (s *ContainerStore) DeleteContainer(ctx context.Context, id string) error {
	return s.etcdClient.Delete(ctx, s.key(id))
}

// ListContainers is used to list all container metadata in etcd.
func (s *ContainerStore) ListContainers(
	ctx context.Context,
) ([]*object.Container, error) {
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

// WatchContainers is used to watch for changes to container metadata in etcd.
func (s *ContainerStore) WatchContainers(
	ctx context.Context,
) (chan *object.Container, error) {
	watchChan := s.etcdClient.cli.Watch(
		ctx,
		containerPrefix,
		clientv3.WithPrefix(),
	)
	containerChan := make(chan *object.Container)

	go s.handleWatchEvents(watchChan, containerChan)

	return containerChan, nil
}

func (s *ContainerStore) handleWatchEvents(
	watchChan clientv3.WatchChan,
	containerChan chan *object.Container,
) {
	defer close(containerChan)

	for resp := range watchChan {
		s.processWatchResponse(resp, containerChan)
	}
}

func (*ContainerStore) processWatchResponse(
	resp clientv3.WatchResponse,
	containerChan chan *object.Container,
) {
	for _, ev := range resp.Events {
		if ev.Type == clientv3.EventTypePut {
			container, err := object.ContainerFromJSON(
				string(ev.Kv.Value),
			)
			if err == nil {
				containerChan <- container
			}
		}
	}
}
