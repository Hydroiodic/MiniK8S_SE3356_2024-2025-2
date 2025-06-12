package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

type HpaStore struct {
	etcdClient *etcd.Client
}

func NewHpaStore(etcdEndpoints []string) (*HpaStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &HpaStore{etcdClient: client}, nil
}

func (s *HpaStore) key(namespace, name string) string {
	return path.Join(etcd.HpaPrefix, namespace, name)
}
func (s *HpaStore) GetHpa(
	ctx context.Context,
	namespace, name string,
) (*HorizontalPodAutoscaler, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var hpa HorizontalPodAutoscaler

	err = json.Unmarshal([]byte(data), &hpa)
	if err != nil {
		return nil, err
	}

	return &hpa, nil
}

func (s *HpaStore) Close() error {
	return s.etcdClient.Close()
}

func (s *HpaStore) AddHpa(
	ctx context.Context,
	hpa *HorizontalPodAutoscaler,
) error {
	jsonData, err := json.Marshal(hpa)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(
		ctx,
		s.key(hpa.Metadata.Namespace, hpa.Metadata.Name),
		string(jsonData),
	)
}

func (s *HpaStore) ListHpas(
	ctx context.Context,
) ([]*HorizontalPodAutoscaler, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.HpaPrefix)
	if err != nil {
		return nil, err
	}

	var hpas []*HorizontalPodAutoscaler

	for _, v := range kvs {
		var hpa HorizontalPodAutoscaler

		err := json.Unmarshal([]byte(v), &hpa)
		if err != nil {
			return nil, err
		}

		hpas = append(hpas, &hpa)
	}

	return hpas, nil
}

func (s *HpaStore) DeleteHpa(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
}
