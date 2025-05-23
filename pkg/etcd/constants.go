// Package etcd provides constants and utilities for
// interacting with etcd in the MiniK8S project.
package etcd

import "time"

const (
	PodPrefix           = "/minik8s/pods/"
	KubeletPrefix       = "/minik8s/kubelets/"
	ResourcePrefix      = "/minik8s/resources/"
	ClusterIPPrefix     = "/minik8s/inner/cluster-ip/"
	DNSPrefix           = "/minik8s/dns/"
	ServicePrefix       = "/minik8s/services/"
	defaultEtcdEndpoint = "localhost:2379"
	defaultTimeout      = 5 * time.Second
	emptyLength         = 0
)
