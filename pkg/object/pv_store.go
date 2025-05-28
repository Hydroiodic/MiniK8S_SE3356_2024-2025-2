package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

// PersistentVolumeStore handles storage operations for PersistentVolume
type PersistentVolumeStore struct {
	etcdClient *etcd.Client
}

// NewPersistentVolumeStore creates a new PersistentVolumeStore
func NewPersistentVolumeStore(
	etcdEndpoints []string,
) (*PersistentVolumeStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}
	return &PersistentVolumeStore{etcdClient: client}, nil
}

// Close closes the etcd client connection
func (s *PersistentVolumeStore) Close() error {
	return s.etcdClient.Close()
}

// key generates the etcd key for a PersistentVolume
func (s *PersistentVolumeStore) key(namespace, name string) string {
	return path.Join(etcd.PersistentVolumePrefix, namespace, name)
}

// ListPersistentVolumes lists all PersistentVolumes
func (s *PersistentVolumeStore) ListPersistentVolumes(
	ctx context.Context,
) ([]*PersistentVolume, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.PersistentVolumePrefix)
	if err != nil {
		return nil, err
	}

	var pvs []*PersistentVolume
	for _, v := range kvs {
		var pv PersistentVolume
		err := json.Unmarshal([]byte(v), &pv)
		if err != nil {
			return nil, err
		}
		pvs = append(pvs, &pv)
	}
	return pvs, nil
}

// GetPersistentVolume retrieves a specific PersistentVolume
func (s *PersistentVolumeStore) GetPersistentVolume(
	ctx context.Context,
	namespace, name string,
) (*PersistentVolume, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}

	var pv PersistentVolume
	err = json.Unmarshal([]byte(data), &pv)
	if err != nil {
		return nil, err
	}
	return &pv, nil
}

// AddPersistentVolume adds a new PersistentVolume
func (s *PersistentVolumeStore) AddPersistentVolume(
	ctx context.Context,
	pv *PersistentVolume,
) error {
	jsonData, err := json.Marshal(pv)
	if err != nil {
		return err
	}
	return s.etcdClient.Put(
		ctx,
		s.key(pv.Metadata.Namespace, pv.Metadata.Name),
		string(jsonData),
	)
}

// DeletePersistentVolume deletes a PersistentVolume
func (s *PersistentVolumeStore) DeletePersistentVolume(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
}

// UpdatePersistentVolume updates an existing PersistentVolume
func (s *PersistentVolumeStore) UpdatePersistentVolume(
	ctx context.Context,
	pv *PersistentVolume,
) error {
	return s.AddPersistentVolume(ctx, pv)
}

// PersistentVolumeClaimStore handles storage operations for PersistentVolumeClaim
type PersistentVolumeClaimStore struct {
	etcdClient *etcd.Client
}

// NewPersistentVolumeClaimStore creates a new PersistentVolumeClaimStore
func NewPersistentVolumeClaimStore(
	etcdEndpoints []string,
) (*PersistentVolumeClaimStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}
	return &PersistentVolumeClaimStore{etcdClient: client}, nil
}

// Close closes the etcd client connection
func (s *PersistentVolumeClaimStore) Close() error {
	return s.etcdClient.Close()
}

// key generates the etcd key for a PersistentVolumeClaim
func (s *PersistentVolumeClaimStore) key(namespace, name string) string {
	return path.Join(etcd.PersistentVolumeClaimPrefix, namespace, name)
}

// ListPersistentVolumeClaims lists all PersistentVolumeClaims
func (s *PersistentVolumeClaimStore) ListPersistentVolumeClaims(
	ctx context.Context,
) ([]*PersistentVolumeClaim, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.PersistentVolumeClaimPrefix)
	if err != nil {
		return nil, err
	}

	var pvcs []*PersistentVolumeClaim
	for _, v := range kvs {
		var pvc PersistentVolumeClaim
		err := json.Unmarshal([]byte(v), &pvc)
		if err != nil {
			return nil, err
		}
		pvcs = append(pvcs, &pvc)
	}
	return pvcs, nil
}

// GetPersistentVolumeClaim retrieves a specific PersistentVolumeClaim
func (s *PersistentVolumeClaimStore) GetPersistentVolumeClaim(
	ctx context.Context,
	namespace, name string,
) (*PersistentVolumeClaim, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}

	var pvc PersistentVolumeClaim
	err = json.Unmarshal([]byte(data), &pvc)
	if err != nil {
		return nil, err
	}
	return &pvc, nil
}

// AddPersistentVolumeClaim adds a new PersistentVolumeClaim
func (s *PersistentVolumeClaimStore) AddPersistentVolumeClaim(
	ctx context.Context,
	pvc *PersistentVolumeClaim,
) error {
	jsonData, err := json.Marshal(pvc)
	if err != nil {
		return err
	}
	return s.etcdClient.Put(
		ctx,
		s.key(pvc.Metadata.Namespace, pvc.Metadata.Name),
		string(jsonData),
	)
}

// DeletePersistentVolumeClaim deletes a PersistentVolumeClaim
func (s *PersistentVolumeClaimStore) DeletePersistentVolumeClaim(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
}

// UpdatePersistentVolumeClaim updates an existing PersistentVolumeClaim
func (s *PersistentVolumeClaimStore) UpdatePersistentVolumeClaim(
	ctx context.Context,
	pvc *PersistentVolumeClaim,
) error {
	return s.AddPersistentVolumeClaim(ctx, pvc)
}
