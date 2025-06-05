package object

import (
	"context"
	"encoding/json"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
)

type EventStore struct {
	etcdClient *etcd.Client
}

// NewDNSStore creates a new DNSStore with the given etcd endpoints.
func NewEventStore(etcdEndpoints []string) (*EventStore, error) {
	client, err := etcd.NewEtcdClient(etcdEndpoints)
	if err != nil {
		return nil, err
	}

	return &EventStore{etcdClient: client}, nil
}

func (s *EventStore) Close() error {
	return s.etcdClient.Close()
}

func (s *EventStore) ListEvents(
	ctx context.Context,
) ([]*Event, error) {
	kvs, err := s.etcdClient.List(ctx, etcd.EventPrefix)
	if err != nil {
		return nil, err
	}

	var es []*Event

	for _, v := range kvs {
		var e Event

		err := json.Unmarshal([]byte(v), &e)
		if err != nil {
			return nil, err
		}

		es = append(es, &e)
	}

	return es, nil
}

func (s *EventStore) key(namespace, name string) string {
	return path.Join(etcd.EventPrefix, namespace, name)
}

func (s *EventStore) GetEvents(
	ctx context.Context,
	namespace, name string,
) (*Function, error) {
	data, err := s.etcdClient.Get(ctx, s.key(namespace, name))
	if err != nil {
		return nil, err
	} else if data == "" {
		return nil, nil
	}

	var e Function

	err = json.Unmarshal([]byte(data), &e)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func (s *EventStore) AddEvent(
	ctx context.Context,
	e *Event,
) error {
	jsonData, err := json.Marshal(e)
	if err != nil {
		return err
	}

	return s.etcdClient.Put(
		ctx,
		s.key(e.Metadata.Namespace, e.Metadata.Name),
		string(jsonData),
	)
}

func (s *EventStore) DeleteEvent(
	ctx context.Context,
	namespace, name string,
) error {
	return s.etcdClient.Delete(ctx, s.key(namespace, name))
}
