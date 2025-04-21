// Package etcd provides constants and utilities for
// interacting with etcd in the MiniK8S project.
package etcd

import "time"

const (
	containerPrefix     = "/minik8s/containers/"
	defaultEtcdEndpoint = "localhost:2379"
	defaultTimeout      = 5 * time.Second
	emptyLength         = 0
)
