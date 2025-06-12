package interfaces

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

//nolint:dupl
func CreateHpa(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var hpa object.HorizontalPodAutoscaler
	if err := c.BindJSON(&hpa); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create HpaStore and check for errors.
	st, err := object.NewHpaStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create HpaStore: "+err.Error(),
		)

		return
	}

	// Ensure the HpaStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close HpaStore: %v\n", closeErr)
		}
	}()

	// Check if the HPA already exists in etcd.
	reply, err := st.GetHpa(
		c.Request.Context(),
		hpa.Metadata.Namespace,
		hpa.Metadata.Name,
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get HPA from etcd: "+err.Error(),
		)

		return
	}

	// If the HPA already exists, return an error.
	if reply != nil {
		c.JSON(
			http.StatusConflict,
			"Create HPA from file failed: same hpa namespace & name",
		)

		return
	}

	if err := st.AddHpa(c.Request.Context(), &hpa); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add HPA to etcd: "+err.Error(),
		)

		return
	}
}

func GetHpas(c *gin.Context) {
	// Create HpaStore and check for errors.
	st, err := object.NewHpaStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create HpaStore: "+err.Error(),
		)

		return
	}

	// Ensure the HpaStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close HpaStore: %v\n", closeErr)
		}
	}()

	// List all HPA in etcd.
	hpas, err := st.ListHpas(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list HPAs from etcd: "+err.Error(),
		)

		return
	}

	// Convert the HPA to a JSON format and return them.
	c.JSON(http.StatusOK, hpas)
}

//nolint:dupl
func DeleteHpa(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var f object.HorizontalPodAutoscaler
	if err := c.BindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create ReplicasetStore and check for errors.
	st, err := object.NewHpaStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create Hpa store: "+err.Error(),
		)

		return
	}

	// Ensure the ReplicasetStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close Hpa store: %v\n", closeErr)
		}
	}()

	// Delete the replicaset from etcd.
	if err := st.DeleteHpa(
		c.Request.Context(),
		f.Metadata.Namespace,
		f.Metadata.Name,
	); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete Hpa from etcd: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "hpa deleted: "+f.Metadata.Name)
}
