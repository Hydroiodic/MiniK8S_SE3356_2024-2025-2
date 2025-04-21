package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultEtcdEndpoint = "localhost:2379"
	defaultTimeout      = 5 * time.Second
)

type EtcdClient struct {
	cli *clientv3.Client
}

func NewEtcdClient(endpoints []string) (*EtcdClient, error) {
	if len(endpoints) == 0 {
		endpoints = []string{defaultEtcdEndpoint}
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: defaultTimeout,
	})
	if err != nil {
		return nil, err
	}

	return &EtcdClient{cli: cli}, nil
}

func (c *EtcdClient) Close() error {
	return c.cli.Close()
}

// Put 存储键值对
func (c *EtcdClient) Put(ctx context.Context, key, value string) error {
	_, err := c.cli.Put(ctx, key, value)
	return err
}

// Get 获取键值
func (c *EtcdClient) Get(ctx context.Context, key string) (string, error) {
	resp, err := c.cli.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return string(resp.Kvs[0].Value), nil
}

// Delete 删除键
func (c *EtcdClient) Delete(ctx context.Context, key string) error {
	_, err := c.cli.Delete(ctx, key)
	return err
}

// List 列出前缀下的所有键值
func (c *EtcdClient) List(ctx context.Context, prefix string) (map[string]string, error) {
	resp, err := c.cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}

	return result, nil
}
