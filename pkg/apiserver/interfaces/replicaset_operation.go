package interfaces

import (
	"fmt"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func CreateReplicaset(c *gin.Context) {
	// Parse the JSON body into a Replicaset object.
	var replicaset object.ReplicaSet
	if err := c.BindJSON(&replicaset); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Check for the namespace and name in the replicaset configuration.
	if replicaset.Metadata.Namespace == "" {
		replicaset.Metadata.Namespace = DefaultNamespace
	}

	// Create ReplicasetStore and check for errors.
	st, err := object.NewReplicasetStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create replicaset store: "+err.Error(),
		)

		return
	}

	// Ensure the ReplicasetStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close replicaset store: %v\n", closeErr)
		}
	}()

	// Check if the replicaset already exists in etcd.
	reply, err := st.GetReplicaset(
		c.Request.Context(),
		replicaset.Metadata.Namespace,
		replicaset.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get replicaset from etcd: "+err.Error(),
		)

		return
	}

	// If the replicaset already exists, return an error.
	if reply != nil {
		fmt.Println(
			"Create ReplicaSet from file failed: same namespace & name",
		)
		c.JSON(
			http.StatusConflict,
			"Create ReplicaSet from file failed: same namespace & name",
		)

		return
	}

	// 写入etcd即可
	// Add the replicaset to etcd.
	if err := st.AddReplicaset(c.Request.Context(), &replicaset); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add replicaset to etcd: "+err.Error(),
		)

		return
	}
}

func UpdateReplicaset(c *gin.Context) {
	// Parse the JSON body into a ReplicaSet object.
	var replicaset object.ReplicaSet
	if err := c.BindJSON(&replicaset); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Check for the namespace and name in the replicaset configuration.
	if replicaset.Metadata.Namespace == "" {
		replicaset.Metadata.Namespace = DefaultNamespace
	}

	// Create ReplicasetStore and check for errors.
	st, err := object.NewReplicasetStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create replicaset store: "+err.Error(),
		)

		return
	}

	// Ensure the ReplicasetStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close replicaset store: %v\n", closeErr)
		}
	}()

	// Check if the replicaset already exists in etcd.
	err = st.UpdateReplicaset(
		c.Request.Context(),
		&replicaset,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get replicaset from etcd: "+err.Error(),
		)

		return
	}
}

func GetReplicasets(c *gin.Context) {
	// Create ReplicasetStore and check for errors.
	st, err := object.NewReplicasetStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create replicaset store: "+err.Error(),
		)

		return
	}

	// Ensure the ReplicasetStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close replicaset store: %v\n", closeErr)
		}
	}()

	// List all replicasets in etcd.
	replicasets, err := st.ListReplicasets(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list replicasets from etcd: "+err.Error(),
		)

		return
	}

	// Convert the replicasets to a JSON format and return them.
	c.JSON(http.StatusOK, replicasets)
}

func DeleteReplicasetFromEtcd(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var rs object.ReplicaSet
	if err := c.BindJSON(&rs); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create ReplicasetStore and check for errors.
	st, err := object.NewReplicasetStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create replicaset store: "+err.Error(),
		)

		return
	}

	// Ensure the ReplicasetStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close replicaset store: %v\n", closeErr)
		}
	}()

	// Delete the replicaset from etcd.
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
