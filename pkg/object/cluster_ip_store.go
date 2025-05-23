package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// ClusterIPStore 提供对 ClusterIP 的增删改查
type ClusterIPStore struct {
	etcdClient *etcd.Client
}

// NewClusterIPStore 创建 ClusterIPStore
func NewClusterIPStore(etcdEndpoints []string) (*ClusterIPStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &ClusterIPStore{etcdClient: client}, nil
}

// Close 关闭 Etcd 客户端
func (s *ClusterIPStore) Close() error {
	return s.etcdClient.Close()
}

// key 生成存储路径
func (s *ClusterIPStore) key(cip string) string {
	return path.Join(etcd.ClusterIPPrefix, cip)
}

// SetClusterIP 存储 ClusterIP
func (s *ClusterIPStore) SetClusterIP(
	ctx context.Context,
	clusterIP string,
) error {
	data, err := json.Marshal(clusterIP)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(ctx, s.key(clusterIP), string(data))
}

// GetClusterIP 获取 ClusterIP
func (s *ClusterIPStore) GetClusterIP(
	ctx context.Context,
	cip string,
) (string, error) {
	data, err := s.etcdClient.Get(ctx, s.key(cip))
	if err != nil {
		return "", err
	} else if data == "" {
		return "", nil
	}

	var clusterIP string
	if err := json.Unmarshal([]byte(data), &clusterIP); err != nil {
		return "", err
	}

	return clusterIP, nil
}

// DeleteClusterIP 删除 ClusterIP
func (s *ClusterIPStore) DeleteClusterIP(
	ctx context.Context,
	cip string,
) error {
	return s.etcdClient.Delete(ctx, s.key(cip))
}

// ListClusterIPs 列出所有 ClusterIP
func (s *ClusterIPStore) ListClusterIPs(
	ctx context.Context,
) ([]string, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.ClusterIPPrefix)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, v := range kvs {
		var clusterIP string
		if err := json.Unmarshal([]byte(v), &clusterIP); err != nil {
			return nil, err
		}

		result = append(result, clusterIP)
	}

	return result, nil
}
