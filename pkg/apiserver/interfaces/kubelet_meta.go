package interfaces

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"slices"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func KubeletRegister(c *gin.Context) {
	// Convert the request body to kubelet struct.
	var kubelet object.Kubelet
	if err := c.BindJSON(&kubelet); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Check if the kubelet is already registered.
	if kubelet.Config.Name == "" {
		c.JSON(http.StatusBadRequest, "The kubelet name is empty.")
		return
	}

	// Create a new etcd connection for kubelet registration.
	st, err := object.NewKubeletStore([]string{})
	if err != nil {
		// Failed to create kubelet store, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create kubelet store: "+err.Error(),
		)

		return
	}

	// Ensure the kubelet store is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close kubelet store: %v\n", closeErr)
		}
	}()

	// Try to get the kubelet object from etcd.
	oldKubelet, err := st.GetKubelet(
		c.Request.Context(),
		kubelet.Config.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get kubelet: "+err.Error(),
		)

		return
	}

	if oldKubelet != nil {
		// If the kubelet already exists, use the same pod list.
		kubelet.Pods = oldKubelet.Pods
		// Register kubelet to apiserver, write into etcd.
		log.Printf(
			"Kubelet already exists: %s, use the previous pod list\n",
			kubelet.Config.Name,
		)
	} else {
		// If the kubelet doesn't exist, create an empty pod list.
		// NOTE: this operation is for data sync. We cannot use the existing pod list.
		kubelet.Pods = make([]object.Pod, 0)
		log.Printf(
			"Registering kubelet %s with an empty pod list\n",
			kubelet.Config.Name,
		)
	}

	// Update the kubelet's last update time.
	kubelet.Heartbeat()

	// Write the kubelet object to etcd.
	if err := st.AddKubelet(c.Request.Context(), &kubelet); err != nil {
		// Failed to write kubelet to etcd, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to register kubelet: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Kubelet registered: "+kubelet.Config.Name)
	log.Println("Kubelet registered: ", kubelet.Config.Name)
}

func KubeletHeartbeat(c *gin.Context) {
	// Convert the request body to kubelet struct.
	var kubelet object.Kubelet
	if err := c.BindJSON(&kubelet); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Check if the kubelet is already registered.
	if kubelet.Config.Name == "" {
		c.JSON(http.StatusBadRequest, "The kubelet name is empty.")
		return
	}

	// Update the kubelet's last update time.
	kubelet.Heartbeat()

	// Create a kubelet store for kubelet heartbeat.
	st, err := object.NewKubeletStore([]string{})
	if err != nil {
		// Failed to create kubelet store, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create kubelet store: "+err.Error(),
		)

		return
	}

	// Ensure the kubelet store is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close kubelet store: %v\n", closeErr)
		}
	}()

	// Create service store for kubelet heartbeat.
	svcStore, err := object.NewServiceStore([]string{})
	if err != nil {
		// Failed to create service store, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create service store: "+err.Error(),
		)

		return
	}

	// Ensure the service store is closed after use.
	defer func() {
		if closeErr := svcStore.Close(); closeErr != nil {
			log.Printf("Failed to close service store: %v\n", closeErr)
		}
	}()

	// TODO: check if the kubelet exists in etcd.
	// Get the kubelet object from etcd.
	oldKubelet, err := st.GetKubelet(
		c.Request.Context(),
		kubelet.Config.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get kubelet: "+err.Error(),
		)

		return
	} else if oldKubelet == nil {
		c.JSON(
			http.StatusNotFound,
			"Kubelet not found: "+kubelet.Config.Name,
		)

		return
	}

	// Sync the pods with the kubelet.
	if err := syncKubeletPods(
		c.Request.Context(), st, oldKubelet, &kubelet,
	); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to sync kubelet pods: "+err.Error(),
		)

		return
	}

	// Sync services with the kubelet.
	if err := syncKubeletServices(
		c.Request.Context(), svcStore, &kubelet,
	); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to sync kubelet services: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Kubelet heartbeat: "+kubelet.Config.Name)
}

func findIndex(
	pods []object.Pod,
	pod object.Pod,
) int {
	// Iterate over the pods to find the index of the specified pod.
	for i, p := range pods {
		if p.Metadata.Name == pod.Metadata.Name &&
			p.Metadata.Namespace == pod.Metadata.Namespace {
			return i
		}
	}

	// If the pod is not found, return -1.
	return -1
}

func checkEndpointsSame(
	ep1, ep2 []object.Endpoint,
) bool {
	// Check if the lengths of the endpoints are the same.
	if len(ep1) != len(ep2) {
		return false
	}

	// Iterate over the endpoints to check if they are the same.
	for _, e1 := range ep1 {
		if !slices.Contains(ep2, e1) {
			return false
		}
	}

	return true
}

func syncKubeletServices(
	ctx context.Context,
	st *object.ServiceStore,
	kubelet *object.Kubelet,
) error {
	// Get all valid services from etcd.
	services, err := st.ListServices(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	// Make some lists to store the services to be added and deleted.
	servicesToAdd := make([]object.Service, 0)
	servicesToDelete := make([]object.Service, 0)

	// Check if there's any valid service not in the kubelet.
	for _, svc := range services {
		// Use a bool flag to check if the service is in the kubelet.
		found := false

		// Check if the service is already in the kubelet.
		for _, kubeSvc := range kubelet.Services {
			// If not the same service, continue.
			if svc.Metadata.Name != kubeSvc.Metadata.Name ||
				svc.Metadata.Namespace != kubeSvc.Metadata.Namespace {
				continue
			}

			// We have found the service in the kubelet.
			found = true

			// Compare their endpoints.
			if !checkEndpointsSame(
				svc.Status.Endpoints,
				kubeSvc.Status.Endpoints,
			) {
				// If the endpoints are different, update the kubelet's service.
				servicesToAdd = append(servicesToAdd, *svc)
			}

			break
		}

		// If the service is not found in the kubelet, add it to the list.
		if !found {
			// Add the service to the kubelet's service list.
			servicesToAdd = append(servicesToAdd, *svc)
		}
	}

	// Check if there are any services in the kubelet not in etcd.
	for _, kubeSvc := range kubelet.Services {
		// Use a bool flag to check if the service is in etcd.
		found := false

		// Check if the service is already in etcd.
		for _, svc := range services {
			// If not the same service, continue.
			if kubeSvc.Metadata.Name != svc.Metadata.Name ||
				kubeSvc.Metadata.Namespace != svc.Metadata.Namespace {
				continue
			}

			// We have found the service in etcd.
			found = true

			break
		}

		// If the service is not found in etcd, add it to the list.
		if !found {
			// Add the service to the kubelet's service list.
			servicesToDelete = append(servicesToDelete, kubeSvc)
		}
	}

	// Log the details of the heartbeat.
	log.Printf(
		"Received heartbeat from kubelet %s: %d services, %d to add, %d to delete\n",
		kubelet.Config.Name,
		len(kubelet.Services),
		len(servicesToAdd),
		len(servicesToDelete),
	)
	log.Printf("Services to add: %v\n", retrieveServicesName(servicesToAdd))
	log.Printf(
		"Services to delete: %v\n",
		retrieveServicesName(servicesToDelete),
	)

	// Execute the operations on the services.
	if err := internalAddKubeletServices(
		kubelet.Config.Name, servicesToAdd,
	); err != nil {
		return fmt.Errorf("failed to add kubelet services: %w", err)
	}

	if err := internalDeleteKubeletServices(
		kubelet.Config.Name, servicesToDelete,
	); err != nil {
		return fmt.Errorf("failed to delete kubelet services: %w", err)
	}

	return nil
}

func syncKubeletPods(
	ctx context.Context,
	st *object.KubeletStore,
	oldKubelet, kubelet *object.Kubelet,
) error {
	// Two arrays of pods are compared.
	podsToUpdate := make([]object.Pod, 0)
	podsToDelete := make([]object.Pod, 0)
	podsToAddKubelet := make([]object.Pod, 0)
	podsToDeleteKubelet := make([]object.Pod, 0)

	// Delete the pods not existing in the new kubelet.
	for _, oldPod := range oldKubelet.Pods {
		// Iterate through the new kubelet's pods to check if the old pod exists.
		i := findIndex(kubelet.Pods, oldPod)

		if i < 0 {
			// If the pod is not found, sync status with Kubelet.
			switch oldPod.Status.Phase {
			case object.PodUnknown,
				object.PodCreating,
				object.PodRunning,
				object.PodFailed:
				oldPod.Status.Phase = object.PodCreating
				podsToAddKubelet = append(podsToAddKubelet, oldPod)
				podsToUpdate = append(podsToUpdate, oldPod)
			case object.PodDeleting:
				podsToDelete = append(podsToDelete, oldPod)
			}
		} else {
			// If the pod is found, update it in the new kubelet.
			// TODO: some containers may not need to be updated.
			switch oldPod.Status.Phase {
			case object.PodUnknown,
				object.PodCreating,
				object.PodRunning,
				object.PodFailed:
				podsToUpdate = append(podsToUpdate, kubelet.Pods[i])
			case object.PodDeleting:
				podsToDeleteKubelet = append(podsToDeleteKubelet, oldPod)
			}
		}
	}

	// If there's any pod in the new kubelet that doesn't exist in the old kubelet,
	// add it to the `podsToDeleteKubelet` list.
	for _, newPod := range kubelet.Pods {
		// Iterate through the old kubelet's pods to check if the new pod exists.
		i := findIndex(oldKubelet.Pods, newPod)

		// If the pod is not found, add it to the delete list.
		if i < 0 {
			newPod.Status.Phase = object.PodDeleting
			podsToDeleteKubelet = append(podsToDeleteKubelet, newPod)
		}
	}

	// Use `podsToUpdate` and `podsToDeleteKubelet` as the new pods list.
	kubelet.Pods = append(podsToUpdate, podsToDeleteKubelet...)

	// Log the details of the heartbeat.
	log.Printf(
		"Received heartbeat from kubelet %s: %d pods, %d to update, %d to re-create, %d to delete\n",
		kubelet.Config.Name,
		len(kubelet.Pods),
		len(podsToUpdate),
		len(podsToAddKubelet),
		len(podsToDelete),
	)
	log.Printf("Pods to update: %v\n", retrievePodsName(podsToUpdate))
	log.Printf("Pods to delete: %v\n", retrievePodsName(podsToDelete))
	log.Printf(
		"Pods to add to kubelet: %v\n",
		retrievePodsName(podsToAddKubelet),
	)
	log.Printf(
		"Pods to delete from kubelet: %v\n",
		retrievePodsName(podsToDeleteKubelet),
	)

	// TODO: ensure the three database operations are atomic.

	// Operations on pods in etcd.
	if err := internalDeletePods(ctx, podsToDelete); err != nil {
		return fmt.Errorf("failed to delete pods: %w", err)
	}

	if err := internalUpdatePods(ctx, podsToUpdate); err != nil {
		return fmt.Errorf("failed to update pods: %w", err)
	}

	// Update the kubelet object in etcd.
	if err := st.UpdateKubelet(ctx, kubelet); err != nil {
		return fmt.Errorf("failed to update kubelet: %w", err)
	}

	// Sync the pods with the kubelet.
	if err := internalSyncPods(
		kubelet.Config.Name,
		podsToAddKubelet,
		podsToDeleteKubelet,
	); err != nil {
		// NOTE: this error is not fatal, because we can sync next time.
		log.Printf("Failed to sync pods with kubelet: %v\n", err)
	}

	return nil
}

func internalAddKubeletServices(
	nodeName string,
	services []object.Service,
) error {
	// Iterate through the services to add them to the kubelet.
	for _, svc := range services {
		// Log the service creation.
		log.Printf(
			"Service %s/%s is re-creating on node %s\n",
			svc.Metadata.Namespace,
			svc.Metadata.Name,
			nodeName,
		)

		// Here we use a message queue to process the service creation.
		msg, err := mqtemplate.CreateServiceMessage(svc)
		if err != nil {
			log.Printf("Failed to create service message: %v\n", err)
			continue
		}

		// Send the message to the queue.
		queueName := path.Join(
			mqtemplate.KubeProxyCreateServiceQueue,
			nodeName,
		)
		if err = mqtemplate.SendMessageToQueue(queueName, msg); err != nil {
			log.Printf("Failed to send message to queue: %v\n", err)
			continue
		}
	}

	return nil
}

func internalDeleteKubeletServices(
	nodeName string,
	services []object.Service,
) error {
	// Iterate through the services to delete them from the kubelet.
	for _, svc := range services {
		// Log the service creation.
		log.Printf(
			"Service %s/%s is deleting on node %s\n",
			svc.Metadata.Namespace,
			svc.Metadata.Name,
			nodeName,
		)

		// Here we use a message queue to process the service creation.
		msg, err := mqtemplate.CreateServiceMessage(svc)
		if err != nil {
			log.Printf("Failed to create service message: %v\n", err)
			continue
		}

		// Send the message to the queue.
		queueName := path.Join(
			mqtemplate.KubeProxyDeleteServiceQueue,
			nodeName,
		)
		if err = mqtemplate.SendMessageToQueue(queueName, msg); err != nil {
			log.Printf("Failed to send message to queue: %v\n", err)
			continue
		}
	}

	return nil
}

func internalDeletePods(ctx context.Context, pods []object.Pod) error {
	// Check for valid kubelet and pods.
	if pods == nil {
		return fmt.Errorf("invalid nil kubelet or pods")
	}

	// Nothing to delete if no pods are provided.
	if len(pods) == 0 {
		return nil
	}

	// Create PodStore and check for errors.
	st, err := object.NewPodStore([]string{})
	if err != nil {
		return fmt.Errorf("failed to create pod store: %w", err)
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// Delete the pods from etcd.
	for _, pod := range pods {
		if err := st.DeletePod(
			ctx,
			pod.Metadata.Namespace,
			pod.Metadata.Name,
		); err != nil {
			// If error occurs, log it and continue.
			log.Printf(
				"Failed to delete pod %s: %v\n",
				path.Join(pod.Metadata.Namespace, pod.Metadata.Name),
				err,
			)
		}
	}

	return nil
}

func internalUpdatePods(ctx context.Context, pods []object.Pod) error {
	// Check for valid kubelet and pods.
	if pods == nil {
		return fmt.Errorf("invalid nil kubelet or pods")
	}

	// Nothing to update if no pods are provided.
	if len(pods) == 0 {
		return nil
	}

	// Create PodStore and check for errors.
	st, err := object.NewPodStore([]string{})
	if err != nil {
		return fmt.Errorf("failed to create pod store: %w", err)
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// Update the pods in etcd.
	for _, pod := range pods {
		if err := st.UpdatePod(ctx, &pod); err != nil {
			log.Printf(
				"Failed to update pod %s: %v\n",
				path.Join(pod.Metadata.Namespace, pod.Metadata.Name),
				err,
			)
		}
	}

	return nil
}

func internalSyncPods(
	nodeName string,
	podsToAddKubelet []object.Pod,
	podsToDeleteKubelet []object.Pod,
) error {
	for _, pod := range podsToAddKubelet {
		// Log the pod creation.
		log.Printf(
			"Pod %s/%s is re-creating on node %s\n",
			pod.Metadata.Namespace,
			pod.Metadata.Name,
			nodeName,
		)

		// Here we use a message queue to process the pod creation.
		msg, err := mqtemplate.CreatePodMessage(pod)
		if err != nil {
			log.Printf("Failed to create pod message: %v\n", err)
			continue
		}

		// Send the message to the queue.
		// NOTE: the name of the queue is `KubeletCreatePodQueue/nodeName`.
		queueName := path.Join(
			mqtemplate.KubeletCreatePodQueue,
			nodeName,
		)
		if err = mqtemplate.SendMessageToQueue(queueName, msg); err != nil {
			log.Printf("Failed to send message to queue: %v\n", err)
			continue
		}
	}

	for _, pod := range podsToDeleteKubelet {
		// Log the pod deletion.
		log.Printf(
			"Pod %s/%s is deleting on node %s\n",
			pod.Metadata.Namespace,
			pod.Metadata.Name,
			nodeName,
		)

		// Here we use a message queue to process the pod deletion.
		msg, err := mqtemplate.CreatePodMessage(pod)
		if err != nil {
			log.Printf("Failed to create pod message: %v\n", err)
			continue
		}

		// Send the message to the queue.
		// NOTE: the name of the queue is `KubeletDeletePodQueue/nodeName`.
		queueName := path.Join(
			mqtemplate.KubeletDeletePodQueue,
			nodeName,
		)
		if err = mqtemplate.SendMessageToQueue(queueName, msg); err != nil {
			log.Printf("Failed to send message to queue: %v\n", err)
			continue
		}
	}

	return nil
}
