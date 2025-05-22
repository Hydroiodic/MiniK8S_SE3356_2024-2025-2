package interfaces

import (
	"log"
	"net/http"
	"path"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func AssignPodToNode(c *gin.Context) {
	// Get query parameters from the request url.
	nodeName := c.Query("nodeName")

	// NOTE: Pods should be created here and saved to etcd.
	var pod object.Pod
	if err := c.BindJSON(&pod); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Check for the namespace and name in the pod configuration.
	if pod.Metadata.Namespace == "" {
		pod.Metadata.Namespace = DefaultNamespace
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

	// Create kubelet store and check for errors.
	ks, err := object.NewKubeletStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create kubelet store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore and KubeletStore is closed after use.
	defer func() {
		// Close PodStore.
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close pod store: %v\n", closeErr)
		}
		// Close KubeletStore.
		if closeErr := ks.Close(); closeErr != nil {
			log.Printf("Failed to close kubelet store: %v\n", closeErr)
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
		log.Println("Create pod from file failed: same pod namespace & name")
		c.JSON(
			http.StatusConflict,
			"Create pod from file failed: same pod namespace & name",
		)

		return
	}

	// Get the kubelet object from etcd.
	kubelet, err := ks.GetKubelet(
		c.Request.Context(),
		nodeName,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get kubelet from etcd: "+err.Error(),
		)

		return
	}

	// If the kubelet does not exist, return an error.
	if kubelet == nil {
		c.JSON(
			http.StatusNotFound,
			"Failed to get kubelet from etcd: "+nodeName+" not found",
		)

		return
	}

	// Update the status of the pod to "Creating".
	pod.Status.Phase = object.PodCreating
	pod.Status.StartTime = time.Now()

	// Update the kubelet object with the pod information.
	kubelet.Pods = append(kubelet.Pods, pod)

	// TODO: ensure the two operations below
	// (adding pod and updating kubelet) are atomic.

	// Add the pod to etcd.
	if err := st.AddPod(c.Request.Context(), &pod); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add pod to etcd: "+err.Error(),
		)

		return
	}

	// Update the kubelet object in etcd.
	if err := ks.UpdateKubelet(c.Request.Context(), kubelet); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to update kubelet in etcd: "+err.Error(),
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
	// NOTE: the name of the queue is `KubeletCreatePodQueue/nodeName`.
	queueName := path.Join(
		mqtemplate.KubeletCreatePodQueue,
		nodeName,
	)
	if err = mqtemplate.SendMessageToQueue(queueName, msg); err != nil {
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
		pod.Metadata.Namespace = DefaultNamespace
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
			log.Printf("Failed to close pod store: %v\n", closeErr)
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
		log.Println("Create pod from file failed: same pod namespace & name")
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

func CreateReplicaset(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var replicaset object.ReplicaSet
	if err := c.BindJSON(&replicaset); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Check for the namespace and name in the pod configuration.
	if replicaset.Metadata.Namespace == "" {
		replicaset.Metadata.Namespace = "default"
	}

	// Create PodStore and check for errors.
	st, err := object.NewReplicasetStore([]string{})
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
	reply, err := st.GetReplicaset(
		c.Request.Context(),
		replicaset.Metadata.Namespace,
		replicaset.Metadata.Name,
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
		fmt.Println(
			"Create ReplicaSet from file failed: same pod namespace & name",
		)
		c.JSON(
			http.StatusConflict,
			"Create ReplicaSet  from file failed: same pod namespace & name",
		)

		return
	}

	//写入etcd即可
	// Add the pod to etcd.
	if err := st.AddReplicaset(c.Request.Context(), &replicaset); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add replicaset to etcd: "+err.Error(),
		)

		return
	}

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
			log.Printf("Failed to close pod store: %v\n", closeErr)
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

func GetReplicasets(c *gin.Context) {
	// Create PodStore and check for errors.
	st, err := object.NewReplicasetStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create replicaset store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close replicaset store: %v\n", closeErr)
		}
	}()

	// List all pods in etcd.
	replicasets, err := st.ListReplicasets(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list pods from etcd: "+err.Error(),
		)

		return
	}

	// Convert the pods to a JSON format and return them.
	c.JSON(http.StatusOK, replicasets)
}

func DeletePod(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var pod object.Pod
	if err := c.BindJSON(&pod); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create KubeletStore and check for errors.
	kubeletStore, err := object.NewKubeletStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create kubelet store: "+err.Error(),
		)

		return
	}

	// Ensure the KubeletStore is closed after use.
	defer func() {
		if closeErr := kubeletStore.Close(); closeErr != nil {
			log.Printf("Failed to close kubelet store: %v\n", closeErr)
		}
	}()

	// Create PodStore and check for errors.
	podStore, err := object.NewPodStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := podStore.Close(); closeErr != nil {
			log.Printf("Failed to close pod store: %v\n", closeErr)
		}
	}()

	// List all kubelets in etcd.
	kubelets, err := kubeletStore.ListKubelets(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list kubelets from etcd: "+err.Error(),
		)

		return
	}

	// Find the kubelet that contains the pod to be deleted.
	var (
		kubelet     *object.Kubelet = nil
		podToUpdate *object.Pod     = nil
	)

	for _, k := range kubelets {
		for i := range k.Pods {
			if k.Pods[i].Metadata.Name == pod.Metadata.Name &&
				k.Pods[i].Metadata.Namespace == pod.Metadata.Namespace {
				// Update the status of the pod to "Deleting".
				k.Pods[i].Status.Phase = object.PodDeleting
				// Assign the kubelet to the variable.
				kubelet = k
				podToUpdate = &k.Pods[i]

				break
			}
		}
	}

	// If not found, return an error.
	if kubelet == nil {
		c.JSON(
			http.StatusNotFound,
			"Delete pod failed: pod not found",
		)

		return
	}

	// Update the pod object in etcd.
	if err := podStore.UpdatePod(c.Request.Context(), podToUpdate); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to update pod in etcd: "+err.Error(),
		)

		return
	}

	// Update the kubelet object in etcd.
	// TODO: data race may happen here.
	if err := kubeletStore.UpdateKubelet(c.Request.Context(), kubelet); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to update kubelet in etcd: "+err.Error(),
		)

		return
	}

	// Create a message for deleting the pod.
	msg, err := mqtemplate.CreatePodMessage(*podToUpdate)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create pod message: "+err.Error(),
		)

		return
	}

	// Send a message to the queue to delete the pod.
	// NOTE: the name of the queue is `KubeletDeletePodQueue/nodeName`.
	queueName := path.Join(
		mqtemplate.KubeletDeletePodQueue,
		kubelet.Config.Name,
	)
	if err := mqtemplate.SendMessageToQueue(queueName, msg); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to send message to queue: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Pod deletion requested: "+podToUpdate.Metadata.Name)
}

// func pdateReplicasetInEtcd(c *gin.Context) {
// 	var rs object.ReplicaSet
// 	if err := c.BindJSON(&rs); err != nil {
// 		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
// 		return
// 	}

// }

func DeleteReplicasetFromEtcd(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var rs object.ReplicaSet
	if err := c.BindJSON(&rs); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create PodStore and check for errors.
	st, err := object.NewReplicasetStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create replicaset store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close replicaset store: %v\n", closeErr)
		}
	}()

	// Delete the pod from etcd.
	if err := st.DeleteReplicaset(
		c.Request.Context(),
		rs.Metadata.Namespace,
		rs.Metadata.Name,
	); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete replicaset from etcd: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Replicaset deleted: "+rs.Metadata.Name)
}

func CreateHpa(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var hpa object.HorizontalPodAutoscaler
	if err := c.BindJSON(&hpa); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Check for the namespace and name in the pod configuration.
	if hpa.Metadata.Namespace == "" {
		hpa.Metadata.Namespace = "default"
	}

	// Create PodStore and check for errors.
	st, err := object.NewHpaStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create hpa store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close hpa store: %v\n", closeErr)
		}
	}()

	// Check if the pod already exists in etcd.
	reply, err := st.GetHpa(
		c.Request.Context(),
		hpa.Metadata.Namespace,
		hpa.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get hpa from etcd: "+err.Error(),
		)

		return
	}

	// If the pod already exists, return an error.
	if reply != nil {
		fmt.Println("Create hpa from file failed: same hpa namespace & name")
		c.JSON(
			http.StatusConflict,
			"Create hpa from file failed: same hpa namespace & name",
		)

		return
	}
	//写入etcd即可
	// Add the hpa to etcd.
	if err := st.AddHpa(c.Request.Context(), &hpa); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add hpa to etcd: "+err.Error(),
		)

		return
	}
}

func GetHpas(c *gin.Context) {
	// Create PodStore and check for errors.
	st, err := object.NewHpaStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create hpa store: "+err.Error(),
		)

		return
	}

	// Ensure the PodStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close hpa store: %v\n", closeErr)
		}
	}()

	// List all pods in etcd.
	pods, err := st.ListHpas(c.Request.Context())
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
