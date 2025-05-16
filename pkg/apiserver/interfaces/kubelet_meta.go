package interfaces

import (
	"fmt"
	"net/http"

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

	// Register kubelet to apiserver, write into etcd.
	fmt.Println("Kubelet registering: ", kubelet.Config.Name)
	kubelet.Pods = nil // TODO: is this necessary?

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
			fmt.Printf("Failed to close kubelet store: %v\n", closeErr)
		}
	}()

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
	fmt.Println("Kubelet registered: ", kubelet.Config.Name)
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
			fmt.Printf("Failed to close kubelet store: %v\n", closeErr)
		}
	}()

	// TODO: check if the kubelet exists in etcd.
	// Get the kubelet object from etcd.
	oldKubelet, err := st.GetKubelet(
		c.Request.Context(),
		kubelet.Config.Name,
	)
	if err == nil && oldKubelet != nil && oldKubelet.Pods != nil {
		// If the kubelet exists, update its pods.
		kubelet.Pods = oldKubelet.Pods
	}

	if err := st.UpdateKubelet(c.Request.Context(), &kubelet); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to update kubelet: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Kubelet heartbeat: "+kubelet.Config.Name)
	fmt.Println("Kubelet heartbeat: ", kubelet.Config.Name)
}
