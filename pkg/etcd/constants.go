// Package etcd provides constants and utilities for
// interacting with etcd in the MiniK8S project.
package etcd

import "time"

const (
	PodPrefix           = "/minik8s/pods/"
	KubeletPrefix       = "/minik8s/kubelets/"
	defaultEtcdEndpoint = "localhost:2379"
	defaultTimeout      = 5 * time.Second
	emptyLength         = 0
)
