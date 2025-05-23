package interfaces

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"

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

	// Create a new etcd connection for kubelet heartbeat.
	st, err := object.NewKubeletStore([]string{})
	if err != nil {
		// Failed to create kubelet store, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create kubelet store: "+err.Error(),
		)

		return
	}

	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close kubelet store: %v\n", closeErr)
		}
	}()

	// TODO: check if the kubelet exists in etcd.
	// Get the kubelet object from etcd.
	oldKubelet, err := st.GetKubelet(
		c.Request.Context(),
		kubelet.Config.Name,
	)

	// Two arrays of pods are compared.
	podsToUpdate := make([]object.Pod, 0)
	podsToDelete := make([]object.Pod, 0)
	podsToAddKubelet := make([]object.Pod, 0)
	podsToDeleteKubelet := make([]object.Pod, 0)

	if err == nil && oldKubelet != nil && oldKubelet.Pods != nil {
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
					podsToAddKubelet = append(podsToAddKubelet, oldPod)
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
					podsToDeleteKubelet = append(podsToDeleteKubelet, kubelet.Pods[i])
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
				podsToDeleteKubelet = append(podsToDeleteKubelet, newPod)
			}
		}

		// Use `podsToAddKubelet` and `podsToUpdate` as the new pods list.
		kubelet.Pods = append(podsToAddKubelet, podsToUpdate...)
	}

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
	if err := internalDeletePods(c.Request.Context(), podsToDelete); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete pods: "+err.Error(),
		)

		return
	}

	if err := internalUpdatePods(c.Request.Context(), podsToUpdate); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to update pods: "+err.Error(),
		)

		return
	}

	// Update the kubelet object in etcd.
	if err := st.UpdateKubelet(c.Request.Context(), &kubelet); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to update kubelet: "+err.Error(),
		)

		return
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
