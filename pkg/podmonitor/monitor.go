package podmonitor

import (
	"context"
	"fmt"
	"time"

	hpa "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/HPA"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type PodMonitor struct {
	cadvisorClient *hpa.CAdvisorClient
	interval       time.Duration
	etcdClient     *object.ResourceStore
}

func NewPodMonitor(
	cadvisorHost string,
	cadvisorPort int,
	interval time.Duration,
) (*PodMonitor, error) {
	// Create the etcd client with the default endpoints.
	etcdClient, err := object.NewResourceStore([]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %v", err)
	}

	return &PodMonitor{
		cadvisorClient: hpa.NewCAdvisorClient(cadvisorHost, cadvisorPort),
		interval:       interval,
		etcdClient:     etcdClient,
	}, nil
}

func (m *PodMonitor) StartMonitoring(containerIDs []string) {
	// Start a timer to periodically check the pod's resource usage.
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Loop to check the pod's resource usage at regular intervals.
	for range ticker.C {
		// Iterate over the container IDs and get their stats.
		for _, containerID := range containerIDs {
			// Get the stats for the container.
			stats, err := m.cadvisorClient.ContainerStats(containerID)
			if err != nil {
				fmt.Printf(
					"Error getting stats for container %s: %v\n",
					containerID,
					err,
				)

				continue
			}

			// If no stats are available, skip this container.
			if len(stats.Stats) == 0 {
				fmt.Printf(
					"No stats available for container %s\n",
					containerID,
				)

				continue
			}

			// Update the stats for the container in etcd.
			m.processStats(containerID, &stats.Stats[len(stats.Stats)-1])
		}
	}
}

func (m *PodMonitor) processStats(
	containerID string,
	stats *object.ContainerStat,
) {
	// Update the pod's resource usage in etcd.
	err := m.etcdClient.UpdateContainerResource(
		context.Background(),
		containerID,
		stats,
	)
	if err != nil {
		fmt.Printf(
			"Error updating resource metrics for container %s: %v\n",
			containerID,
			err,
		)

		return
	}
}
