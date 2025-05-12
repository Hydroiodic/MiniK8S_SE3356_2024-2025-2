package interfaces

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func GetNodes(c *gin.Context) {
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

	// Get the list of nodes from the kubelet store.
	nodes, err := st.ListKubelets(c.Request.Context())
	if err != nil {
		// Failed to list kubelets, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list kubelets: "+err.Error(),
		)

		return
	}

	// Serialize the nodes to JSON.
	jsonData, err := json.Marshal(nodes)
	if err != nil {
		// Failed to serialize nodes, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to serialize nodes: "+err.Error(),
		)

		return
	}

	// Send the serialized nodes as a JSON response.
	c.JSON(http.StatusOK, string(jsonData))
}
