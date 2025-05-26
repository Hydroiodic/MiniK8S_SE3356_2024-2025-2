// coroutine_utils.go
// We declare some coroutine interfaces here for status checking.
package interfaces

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func CheckKubeletTimeout() error {
	// Create a context used for the etcd connection.
	ctx := context.Background()

	// Create a new etcd connection for status checking.
	st, err := object.NewKubeletStore([]string{})
	if err != nil {
		// Failed to create kubelet store, report error.
		return err
	}

	// Ensure the kubelet store is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close kubelet store: %v\n", closeErr)
		}
	}()

	// Get all kubelet objects from etcd.
	kubelets, err := st.ListKubelets(ctx)
	if err != nil {
		// Failed to list kubelets, report error.
		return err
	}

	// NOTE: In `kubelet/runtime/status_controller.go`, the interval of heartbeat is 5s.
	//       Here we use three times of that as the timeout.
	timeout := 5 * 5 * time.Second

	// We use an array to store kubelets that have timed out.
	timeout_kubelets := make([]*object.Kubelet, 0)

	// Check if any kubelet has timed out.
	for _, kubelet := range kubelets {
		if kubelet.LastUpdateTime.Add(timeout).Before(time.Now()) {
			// Kubelet has timed out, report error.
			log.Printf("Kubelet %s has timed out\n", kubelet.Config.Name)
			timeout_kubelets = append(timeout_kubelets, kubelet)
		}
	}

	// Remove timed out kubelets from etcd.
	for _, kubelet := range timeout_kubelets {
		// First let's remove all pods in this kubelet.
		// TODO: Shall we assign the pods to other nodes?
		if err := internalDeletePods(ctx, kubelet.Pods); err != nil {
			log.Printf(
				"Failed to delete pods for kubelet %s: %v\n",
				kubelet.Config.Name,
				err,
			)

			continue
		}

		// TODO: ensure transaction is atomic.
		if err := st.DeleteKubelet(ctx, kubelet.Config.Name); err != nil {
			// Failed to delete timed out kubelet, report error.
			log.Printf(
				"Failed to delete timed out kubelet %s: %v\n",
				kubelet.Config.Name,
				err,
			)

			continue
		}

		log.Printf("Deleted timed out kubelet %s\n", kubelet.Config.Name)
	}

	return nil
}

/**
 * The function `SyncEtcd` will work on two tasks:
 *  1. Check pods status:
 *     If pods are not consistent with them in kubelet, delete them.
 *  2. Update service status:
 *     Update endpoints for services by checking labels of pods.
 *	   If the service is ready to be applied to the cluster,
 *     move it from pending to valid.
 */
func SyncEtcd() error {
	// NOTE: Not used now.
	return nil
}

// Maybe this function should run in a STW routine?
// But that hasn't been implemented in our project yet.
// So as I think, the function below should not be called whatsoever.
func SyncEtcdPods() error {
	// Create a context used for the etcd connection.
	ctx := context.Background()

	// Create a new kubelet store for status checking.
	st, err := object.NewKubeletStore([]string{})
	if err != nil {
		// Failed to create kubelet store, report error.
		return err
	}

	// Ensure the kubelet store is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close kubelet store: %v\n", closeErr)
		}
	}()

	// Create a new pod store for status checking.
	podStore, err := object.NewPodStore([]string{})
	if err != nil {
		// Failed to create pod store, report error.
		return err
	}

	// Ensure the pod store is closed after use.
	defer func() {
		if closeErr := podStore.Close(); closeErr != nil {
			log.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// Get all kubelet objects from etcd.
	kubelets, err := st.ListKubelets(ctx)
	if err != nil {
		// Failed to list kubelets, report error.
		return err
	}

	// Get all pods objects from etcd.
	pods, err := podStore.ListPods(ctx)
	if err != nil {
		// Failed to list pods, report error.
		return err
	}

	// Make some lists for later use.
	kubeletToUpdate := make([]*object.Kubelet, 0)

	// Check if any pod only exists in either pods or kubelets.
	for _, kubelet := range kubelets {
		// Make a list to save valid pods for this kubelet.
		validPods := make([]object.Pod, 0)

		for _, kubepod := range kubelet.Pods {
			// Check if the pod exists in etcd.
			for j, pod := range pods {
				if pod.Metadata.Namespace == kubepod.Metadata.Namespace &&
					pod.Metadata.Name == kubepod.Metadata.Name {
					// Pod exists in etcd.
					// Remove the pod from the array `pods`.
					pods = slices.Delete(pods, j, j+1)
					// Add the pod to the list of valid pods.
					validPods = append(validPods, kubepod)

					break
				}
			}

			// The kubelet pod does not exist in pods.
			// We need to update the kubelet.
			// NOTE: This is unlikely to happen.
		} //nolint

		// If the length of two kubelet pods is not equal,
		// we need to update the kubelet.
		if len(kubelet.Pods) != len(validPods) {
			kubelet.Pods = validPods
			kubeletToUpdate = append(kubeletToUpdate, kubelet)
		}
	}

	// Now `KubeletToUpdate` contains all kubelets that need to be updated.
	// And `pods` contains all pods that need to be deleted.
	// NOTE: This is unlikely to happen. So we log here.
	if len(kubeletToUpdate) > 0 {
		log.Fatalf(
			"Sanity check failed: try to update kubelets: %v\n",
			kubeletToUpdate,
		)
	}

	if len(pods) > 0 {
		log.Fatalf(
			"Sanity check failed: try to delete pods: %v\n",
			pods,
		)
	}

	// Now we update the kubelet and delete the pods.
	// TODO: ensure the transaction below to be atomic.
	for _, kubelet := range kubeletToUpdate {
		if err := st.UpdateKubelet(ctx, kubelet); err != nil {
			// Failed to update kubelet, report error.
			log.Printf(
				"Failed to update kubelet %s: %v\n",
				kubelet.Config.Name,
				err,
			)

			continue
		}
	}

	for _, pod := range pods {
		if err := podStore.DeletePod(ctx,
			pod.Metadata.Namespace,
			pod.Metadata.Name); err != nil {
			// Failed to delete pod, report error.
			log.Printf(
				"Failed to delete pod %s: %v\n",
				pod.Metadata.Name,
				err,
			)

			continue
		}
	}

	return nil
}

func SyncEtcdServices() error {
	// Create a context used for the etcd connection.
	ctx := context.Background()

	// Create a new service store for status checking.
	st, err := object.NewServiceStore([]string{})
	if err != nil {
		// Failed to create service store, report error.
		return err
	}

	// Ensure the service store is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close service store: %v\n", closeErr)
		}
	}()

	// Create a new pod store for status checking.
	podStore, err := object.NewPodStore([]string{})
	if err != nil {
		// Failed to create pod store, report error.
		return err
	}

	// Ensure the pod store is closed after use.
	defer func() {
		if closeErr := podStore.Close(); closeErr != nil {
			log.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// Get all ready services objects from etcd.
	services, err := st.ListServicesWithoutStatus(ctx)
	if err != nil {
		// Failed to list services, report error.
		return err
	}

	// Get all pods objects from etcd.
	pods, err := podStore.ListPods(ctx)
	if err != nil {
		// Failed to list pods, report error.
		return err
	}

	// Make some lists for later use.
	servicesToUpdate := make([]*object.Service, 0)

	// Iterate through all services and find their endpoints.
	for _, service := range services {
		// NOTE: Because DNSService and ProxyService are internal services,
		// 	     there're no pods related, so we should skip them.
		if service.Type == object.SERVICE_TYPE_CLUSTERIP_STR {
			continue
		}

		// Make a list to save valid pods for this service.
		endpoints := make([]object.Endpoint, 0)

		// Iterate through all pods and find their endpoints.
		for _, pod := range pods {
			// Only Running pods are considered.
			if pod.Status.Phase != object.PodRunning {
				continue
			}

			// Get all exposed ports of the pod.
			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					endpoints = append(endpoints, object.Endpoint{
						IP:   pod.Status.IP,
						Port: port,
					})
				}
			}
		}

		// Check if the endpoints are different from the service.
		if len(endpoints) != len(service.Status.Endpoints) {
			// Update the service endpoints.
			service.Status.Endpoints = endpoints
			servicesToUpdate = append(servicesToUpdate, service)
		} else {
			// Check if the endpoints are different.
			for _, endpoint := range endpoints {
				if !slices.Contains(service.Status.Endpoints, endpoint) {
					// Update the service endpoints.
					service.Status.Endpoints = endpoints
					servicesToUpdate = append(servicesToUpdate, service)

					break
				}
			}
		}
	}

	// If there's any services that need to be updated,
	// DNS and proxy will be updated later.
	if len(servicesToUpdate) != 0 {
		forwardingInfoMutex.Lock()
		forwardingInfoNeedUpdate = true
		forwardingInfoMutex.Unlock()
	}

	// Update the services.
	for _, service := range servicesToUpdate {
		if err := st.UpdateServiceWithoutStatus(ctx, service); err != nil {
			// Failed to update service, report error.
			log.Printf(
				"Failed to update service %s: %v\n",
				service.Metadata.Name,
				err,
			)

			continue
		}
	}

	return nil
}
