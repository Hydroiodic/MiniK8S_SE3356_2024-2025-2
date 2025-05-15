package interfaces

import (
	"fmt"
	"net/http"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func AssignPodToNode(c *gin.Context) {
	// NOTE: Pods should be created here and saved to etcd.
	var pod object.Pod
	if err := c.BindJSON(&pod); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Check for the namespace and name in the pod configuration.
	if pod.Metadata.Namespace == "" {
		pod.Metadata.Namespace = "default"
	}

	// Create PodStore and check for errors.
	st, err := object.NewPodStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// Check if the pod already exists in etcd.
	reply, err := st.GetPod(
		c.Request.Context(),
		pod.Metadata.Namespace,
		pod.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get pod from etcd: "+err.Error(),
		)

		return
	}

	// If the pod already exists, return an error.
	if reply != nil {
		fmt.Println("Create pod from file failed: same pod namespace & name")
		c.JSON(
			http.StatusConflict,
			"Create pod from file failed: same pod namespace & name",
		)

		return
	}

	// TODO: update the status of the pod to "Creating".

	// Add the pod to etcd.
	if err := st.AddPod(c.Request.Context(), &pod); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add pod to etcd: "+err.Error(),
		)

		return
	}

	// Now we create a message queue to notify the kubelet.
	msg, err := mqtemplate.CreatePodMessage(pod)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod message: "+err.Error(),
		)

		return
	}

	// Send the message to the queue.
	err = mqtemplate.SendMessageToQueue(mqtemplate.KubeletCreatePodQueue, msg)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to send message to queue: "+err.Error(),
		)

		return
	}

	// Return OK response.
	path := path.Join(
		etcd.PodPrefix,
		pod.Metadata.Namespace,
		pod.Metadata.Name,
	)
	c.JSON(http.StatusOK, "Pod created: "+path)
}

func CreatePod(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var pod object.Pod
	if err := c.BindJSON(&pod); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Check for the namespace and name in the pod configuration.
	if pod.Metadata.Namespace == "" {
		pod.Metadata.Namespace = "default"
	}

	// Create PodStore and check for errors.
	st, err := object.NewPodStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// Check if the pod already exists in etcd.
	reply, err := st.GetPod(
		c.Request.Context(),
		pod.Metadata.Namespace,
		pod.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get pod from etcd: "+err.Error(),
		)

		return
	}

	// If the pod already exists, return an error.
	if reply != nil {
		fmt.Println("Create pod from file failed: same pod namespace & name")
		c.JSON(
			http.StatusConflict,
			"Create pod from file failed: same pod namespace & name",
		)

		return
	}

	// Add the pod to etcd.
	// Here we use a message queue to process the pod creation.
	msg, err := mqtemplate.CreatePodMessage(pod)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod message: "+err.Error(),
		)

		return
	}

	// Send the message to the queue.
	err = mqtemplate.SendMessageToQueue(mqtemplate.CreatePodQueueName, msg)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to send message to queue: "+err.Error(),
		)

		return
	}

	// Return OK response.
	path := path.Join(
		etcd.PodPrefix,
		pod.Metadata.Namespace,
		pod.Metadata.Name,
	)
	c.JSON(http.StatusOK, "Creating pod from file: "+path)
}

func GetPods(c *gin.Context) {
	// Create PodStore and check for errors.
	st, err := object.NewPodStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// List all pods in etcd.
	pods, err := st.ListPods(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list pods from etcd: "+err.Error(),
		)

		return
	}

	// Convert the pods to a JSON format and return them.
	c.JSON(http.StatusOK, pods)
}

func DeletePod(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var pod object.Pod
	if err := c.BindJSON(&pod); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Ensure the pod exists in etcd.
	st, err := object.NewPodStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// Check if the pod exists in etcd.
	reply, err := st.GetPod(
		c.Request.Context(),
		pod.Metadata.Namespace,
		pod.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get pod from etcd: "+err.Error(),
		)

		return
	}

	// If the pod does not exist, return an error.
	if reply == nil {
		fmt.Println("Delete pod failed: pod not found")
		c.JSON(
			http.StatusNotFound,
			"Delete pod failed: pod not found",
		)

		return
	}

	// TODO: update the status of the pod to "Deleting".

	// Create a message for deleting the pod.
	msg, err := mqtemplate.CreatePodMessage(pod)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod message: "+err.Error(),
		)

		return
	}

	// Send a message to the queue to delete the pod.
	err = mqtemplate.SendMessageToQueue(
		mqtemplate.KubeletDeletePodQueue,
		msg,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to send message to queue: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Pod deletion requested: "+pod.Metadata.Name)
}

func DeletePodFromEtcd(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var pod object.Pod
	if err := c.BindJSON(&pod); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create PodStore and check for errors.
	st, err := object.NewPodStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// Delete the pod from etcd.
	if err := st.DeletePod(
		c.Request.Context(),
		pod.Metadata.Namespace,
		pod.Metadata.Name,
	); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete pod from etcd: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Pod deleted: "+pod.Metadata.Name)
}
