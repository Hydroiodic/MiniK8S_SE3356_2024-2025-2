package object

import "slices"

type ContainerStats struct {
	Spec  ContainerSpec   `json:"spec"`
	Stats []ContainerStat `json:"stats"`
}

type ContainerSpec struct {
	CreationTime string            `json:"creation_time"`
	Labels       map[string]string `json:"labels"`
	HasCPU       bool              `json:"has_cpu"`
	HasMemory    bool              `json:"has_memory"`
	HasNetwork   bool              `json:"has_network"`
}

type ContainerStat struct {
	Timestamp string `json:"timestamp"`
	CPU       struct {
		Usage struct {
			Total  uint64   `json:"total"`
			PerCPU []uint64 `json:"per_cpu_usage,omitempty"`
			User   uint64   `json:"user"`
			System uint64   `json:"system"`
		} `json:"usage"`
	} `json:"cpu"`
	Memory struct {
		Usage      uint64 `json:"usage"`
		WorkingSet uint64 `json:"working_set"`
		Cache      uint64 `json:"cache"`
		RSS        uint64 `json:"rss"`
	} `json:"memory"`
	Network struct {
		Interfaces []InterfaceStats `json:"interfaces,omitempty"`
	} `json:"network,omitempty"`
}

type InterfaceStats struct {
	Name      string `json:"name"`
	RxBytes   uint64 `json:"rx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
}

// PodResourceMetrics is used to store resource metrics in etcd.
type PodResourceMetrics struct {
	Namespace string `json:"namespace"`
	PodName   string `json:"podName"`
	Timestamp string `json:"timestamp"`
	CPU       struct {
		UsageNanoCores       uint64 `json:"usageNanoCores"`
		UsageCoreNanoSeconds uint64 `json:"usageCoreNanoSeconds"`
	} `json:"cpu"`
	Memory struct {
		UsageBytes      uint64 `json:"usageBytes"`
		WorkingSetBytes uint64 `json:"workingSetBytes"`
		RSSBytes        uint64 `json:"rssBytes"`
	} `json:"memory"`
	Network struct {
		RxBytes uint64 `json:"rxBytes"`
		TxBytes uint64 `json:"txBytes"`
	} `json:"network"`
}

func (s *ContainerStat) ToPodResourceMetrics(
	namespace, podName string,
) *PodResourceMetrics {
	return &PodResourceMetrics{
		Namespace: namespace,
		PodName:   podName,
		Timestamp: s.Timestamp,
		CPU: struct {
			UsageNanoCores       uint64 `json:"usageNanoCores"`
			UsageCoreNanoSeconds uint64 `json:"usageCoreNanoSeconds"`
		}{
			UsageNanoCores:       s.CPU.Usage.Total,
			UsageCoreNanoSeconds: s.CPU.Usage.Total,
		},
		Memory: struct {
			UsageBytes      uint64 `json:"usageBytes"`
			WorkingSetBytes uint64 `json:"workingSetBytes"`
			RSSBytes        uint64 `json:"rssBytes"`
		}{
			UsageBytes:      s.Memory.Usage,
			WorkingSetBytes: s.Memory.WorkingSet,
			RSSBytes:        s.Memory.RSS,
		},
	}
}

func (s *ContainerStat) CombinedWith(other *ContainerStat) *ContainerStat {
	return &ContainerStat{
		Timestamp: s.Timestamp,
		CPU: struct {
			Usage struct {
				Total  uint64   `json:"total"`
				PerCPU []uint64 `json:"per_cpu_usage,omitempty"`
				User   uint64   `json:"user"`
				System uint64   `json:"system"`
			} `json:"usage"`
		}{
			Usage: struct {
				Total  uint64   `json:"total"`
				PerCPU []uint64 `json:"per_cpu_usage,omitempty"`
				User   uint64   `json:"user"`
				System uint64   `json:"system"`
			}{
				Total:  s.CPU.Usage.Total + other.CPU.Usage.Total,
				User:   s.CPU.Usage.User + other.CPU.Usage.User,
				System: s.CPU.Usage.System + other.CPU.Usage.System,
			},
		},
		Memory: struct {
			Usage      uint64 `json:"usage"`
			WorkingSet uint64 `json:"working_set"`
			Cache      uint64 `json:"cache"`
			RSS        uint64 `json:"rss"`
		}{
			Usage:      s.Memory.Usage + other.Memory.Usage,
			WorkingSet: s.Memory.WorkingSet + other.Memory.WorkingSet,
			Cache:      s.Memory.Cache + other.Memory.Cache,
			RSS:        s.Memory.RSS + other.Memory.RSS,
		},
		Network: struct {
			Interfaces []InterfaceStats `json:"interfaces,omitempty"`
		}{
			Interfaces: slices.Concat(
				s.Network.Interfaces,
				other.Network.Interfaces,
			),
		},
	}
}
