// coroutine_utils.go
// We declare some coroutine interfaces here for status checking.
package interfaces

import (
	"context"
	"fmt"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func CheckKubeletTimeout() error {
	// Create a new etcd connection for status checking.
	st, err := object.NewKubeletStore([]string{})
	if err != nil {
		// Failed to create kubelet store, report error.
		return err
	}

	// Ensure the kubelet store is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close kubelet store: %v\n", closeErr)
		}
	}()

	// Get all kubelet objects from etcd.
	kubelets, err := st.ListKubelets(context.Background())
	if err != nil {
		// Failed to list kubelets, report error.
		return err
	}

	// NOTE: In `kubelet/runtime/status_controller.go`, the interval of heartbeat is 10s.
	//       Here we use three times of that as the timeout.
	timeout := 3 * 10 * time.Second

	// We use an array to store kubelets that have timed out.
	timeout_kubelets := make([]*object.Kubelet, 0)

	// Check if any kubelet has timed out.
	for _, kubelet := range kubelets {
		if kubelet.LastUpdateTime.Add(timeout).Before(time.Now()) {
			// Kubelet has timed out, report error.
			fmt.Printf("Kubelet %s has timed out\n", kubelet.Config.Name)
			timeout_kubelets = append(timeout_kubelets, kubelet)
		}
	}

	// Remove timed out kubelets from etcd.
	for _, kubelet := range timeout_kubelets {
		if err := st.DeleteKubelet(context.Background(), kubelet.Config.Name); err != nil {
			// Failed to delete timed out kubelet, report error.
			fmt.Printf(
				"Failed to delete timed out kubelet %s: %v\n",
				kubelet.Config.Name,
				err,
			)
		} else {
			fmt.Printf("Deleted timed out kubelet %s\n", kubelet.Config.Name)
		}
	}

	return nil
}
