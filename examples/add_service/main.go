package main

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/ipvs_ops"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func main() {
	// 初始化 IPVS 操作对象
	ops := ipvs_ops.NewIpvsOps(
		ipvs_ops.CLUSTER_CIDR_DEFAULT,
	) // 假设 ClusterIPCIDR 为 "222.111.0.0/16"
	defer ops.Close()
	ops.Clear()
	ops.Init()

	// 2. 创建一个ClusterIP类型的Service
	svc := &object.Service{
		Kind: "Service",
		Type: object.SERVICE_TYPE_CLUSTERIP_STR,
		Metadata: object.Metadata{
			Name:      "test-clusterip",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: object.ServiceSpec{
			Selector: map[string]string{"app": "test"},
			Ports: []object.ServicePort{
				{
					Name:       "http",
					Port:       53,
					TargetPort: 5300,
				},
			},
		},
		Status: object.ServiceStatus{
			ClusterIP: "222.111.0.100", // 假设 ClusterIP
			Endpoints: []object.Endpoint{
				{
					IP:   "192.168.1.6", // 必须是节点的IP地址
					Port: 5300,
				},
			},
		},
	}

	ops.AddService(svc)
}
