package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

type GPUJobStore struct {
	etcdClient *etcd.Client
}

// NewDNSStore creates a new DNSStore with the given etcd endpoints.
func NewGPUJobStore(etcdEndpoints []string) (*GPUJobStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &GPUJobStore{etcdClient: client}, nil
}

func (s *GPUJobStore) Close() error {
	return s.etcdClient.Close()
}

func (s *GPUJobStore) ListGPUJobs(
	ctx context.Context,
) ([]*Job, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.GPUJobPrefix)
	if err != nil {
		return nil, err
	}

	var jobs []*Job

	for _, v := range kvs {
		var job Job

		err := json.Unmarshal([]byte(v), &job)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (s *GPUJobStore) key(namespace, name string) string {
	return path.Join(etcd.GPUJobPrefix, namespace, name)
}

func (s *GPUJobStore) GetGPUJob(
	ctx context.Context,
	namespace, name string,
) (*Job, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var job Job

	err = json.Unmarshal([]byte(data), &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *GPUJobStore) AddGPUJob(
	ctx context.Context,
	job *Job,
) error {
	jsonData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(
		ctx,
		s.key(job.Metadata.Namespace, job.Metadata.Name),
		string(jsonData),
	)
}
