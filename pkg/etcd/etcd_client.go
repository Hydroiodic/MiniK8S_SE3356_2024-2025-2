package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Client provides a simple client for interacting with an etcd cluster.
type Client struct {
	cli *clientv3.Client
}

// NewEtcdClient creates a new etcd client with the given endpoints.
func NewEtcdClient(endpoints []string) (*Client, error) {
	if len(endpoints) == emptyLength {
		endpoints = []string{defaultEtcdEndpoint}
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: defaultTimeout,
	})
	if err != nil {
		return nil, err
	}

	return &Client{cli: cli}, nil
}

// Close is used to close the etcd client connection.
func (c *Client) Close() error {
	return c.cli.Close()
}

// Put is used to store a key-value pair in etcd.
func (c *Client) Put(ctx context.Context, key, value string) error {
	_, err := c.cli.Put(ctx, key, value)
	return err
}

// Get is used to retrieve the value of a key from etcd.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	resp, err := c.cli.Get(ctx, key)

	// If any error occurs, return the error.
	if err != nil {
		return "", err
	} else if len(resp.Kvs) == emptyLength {
		return "", nil
	}

	return string(resp.Kvs[0].Value), nil
}

// Delete is used to delete a key from etcd.
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.cli.Delete(ctx, key)
	return err
}

// List is used to list all keys with a given prefix from etcd.
func (c *Client) List(
	ctx context.Context,
	prefix string,
) (map[string]string, error) {
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
