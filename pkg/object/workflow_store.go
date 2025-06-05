package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

type WorkflowStore struct {
	etcdClient *etcd.Client
}

// NewDNSStore creates a new DNSStore with the given etcd endpoints.
func NewWorkflowStore(etcdEndpoints []string) (*WorkflowStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &WorkflowStore{etcdClient: client}, nil
}

func (s *WorkflowStore) Close() error {
	return s.etcdClient.Close()
}

func (s *WorkflowStore) ListWorkflows(
	ctx context.Context,
) ([]*Workflow, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.WorkflowPrefix)
	if err != nil {
		return nil, err
	}

	var wfs []*Workflow

	for _, v := range kvs {
		var wf Workflow

		err := json.Unmarshal([]byte(v), &wf)
		if err != nil {
			return nil, err
		}

		wfs = append(wfs, &wf)
	}

	return wfs, nil
}

func (s *WorkflowStore) key(namespace, name string) string {
	return path.Join(etcd.WorkflowPrefix, namespace, name)
}

func (s *WorkflowStore) GetWorkflows(
	ctx context.Context,
	namespace, name string,
) (*Workflow, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var wf Workflow

	err = json.Unmarshal([]byte(data), &wf)
	if err != nil {
		return nil, err
	}

	return &wf, nil
}

func (s *WorkflowStore) AddWorkflow(
	ctx context.Context,
	fc *Workflow,
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

func (s *WorkflowStore) DeleteWorkflow(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
}
