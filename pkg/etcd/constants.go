// Package etcd provides constants and utilities for
// interacting with etcd in the MiniK8S project.
package etcd

import "time"

const (
	PodPrefix        = "/minik8s/pods/"
	KubeletPrefix    = "/minik8s/kubelets/"
	ResourcePrefix   = "/minik8s/resources/"
	ClusterIPPrefix  = "/minik8s/clusterip/"
	DNSPrefix        = "/minik8s/dns/"
	ReplicasetPrefix = "/minik8s/replicaset"
	HpaPrefix        = "/minik8s/hpa"
	GpujobPrefix     = "/minik8s/gpujob"

	// The services in etcd will diff in the following two states:
	// 1. Pending: The service is created but not yet applied to the cluster.
	//             Because cluster IP or endpoint has not been assigned.
	// 2. Valid: The service is created and applied to the cluster.
	PendingServicePrefix = "/minik8s/services/pending/"
	ValidServicePrefix   = "/minik8s/services/valid/"
	AllServicePrefix     = "/minik8s/services/"

	defaultEtcdEndpoint = "localhost:2379"
	defaultTimeout      = 5 * time.Second
	emptyLength         = 0
)
