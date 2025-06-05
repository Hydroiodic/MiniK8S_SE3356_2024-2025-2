package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

type FuncStore struct {
	etcdClient *etcd.Client
}

// NewDNSStore creates a new DNSStore with the given etcd endpoints.
func NewFuncStore(etcdEndpoints []string) (*FuncStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &FuncStore{etcdClient: client}, nil
}

func (s *FuncStore) Close() error {
	return s.etcdClient.Close()
}

func (s *FuncStore) ListFunctions(
	ctx context.Context,
) ([]*Function, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.FunctionPrefix)
	if err != nil {
		return nil, err
	}

	var funcs []*Function

	for _, v := range kvs {
		var fc Function

		err := json.Unmarshal([]byte(v), &fc)
		if err != nil {
			return nil, err
		}

		funcs = append(funcs, &fc)
	}

	return funcs, nil
}

func (s *FuncStore) key(namespace, name string) string {
	return path.Join(etcd.FunctionPrefix, namespace, name)
}

func (s *FuncStore) GetFuntions(
	ctx context.Context,
	namespace, name string,
) (*Function, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var fc Function

	err = json.Unmarshal([]byte(data), &fc)
	if err != nil {
		return nil, err
	}

	return &fc, nil
}

func (s *FuncStore) AddFunction(
	ctx context.Context,
	fc *Function,
) error {
	jsonData, err := json.Marshal(fc)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(
		ctx,
		s.key(fc.Metadata.Namespace, fc.Metadata.Name),
		string(jsonData),
	)
}

func (s *FuncStore) DeleteFunction(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
}
