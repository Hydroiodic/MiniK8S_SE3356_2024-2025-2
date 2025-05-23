package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

type ReplicasetStore struct {
	etcdClient *etcd.Client
}

func NewReplicasetStore(etcdEndpoints []string) (*ReplicasetStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &ReplicasetStore{etcdClient: client}, nil
}

// Close is used to close the etcd client connection.
func (s *ReplicasetStore) Close() error {
	return s.etcdClient.Close()
}

func (s *ReplicasetStore) ListReplicasets(
	ctx context.Context,
) ([]*ReplicaSet, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.ReplicasetPrefix)
	if err != nil {
		return nil, err
	}

	var replicasets []*ReplicaSet

	for _, v := range kvs {
		var replicaset ReplicaSet

		err := json.Unmarshal([]byte(v), &replicaset)
		if err != nil {
			return nil, err
		}

		replicasets = append(replicasets, &replicaset)
	}

	return replicasets, nil
}

func (s *ReplicasetStore) key(namespace, name string) string {
	return path.Join(etcd.ReplicasetPrefix, namespace, name)
}

func (s *ReplicasetStore) GetReplicaset(
	ctx context.Context,
	namespace, name string,
) (*ReplicaSet, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var replicaset ReplicaSet

	err = json.Unmarshal([]byte(data), &replicaset)
	if err != nil {
		return nil, err
	}

	return &replicaset, nil
}

func (s *ReplicasetStore) AddReplicaset(
	ctx context.Context,
	rs *ReplicaSet,
) error {
	jsonData, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(
		ctx,
		s.key(rs.Metadata.Namespace, rs.Metadata.Name),
		string(jsonData),
	)
}

func (s *ReplicasetStore) DeleteReplicaset(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
}

func (s *ReplicasetStore) UpdateReplicaset(ctx context.Context,
	rs *ReplicaSet,
) error {
	return s.AddReplicaset(ctx, rs)
}
