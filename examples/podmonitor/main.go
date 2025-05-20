package main

import (
	"fmt"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/podmonitor"
)

func main() {
	// Create a new PodMonitor instance.
	podMonitor, err := podmonitor.NewPodMonitor("", 0, 5*time.Second)
	if err != nil {
		fmt.Printf("Error creating PodMonitor: %v\n", err)
		return
	}

	// Start monitoring a specific pod.
	// NOTE: replace the empty strings with actual container IDs.
	podMonitor.StartMonitoring([]string{"", ""})
}
