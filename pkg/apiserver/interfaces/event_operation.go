package interfaces

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func GetEvents(c *gin.Context) {
	// Create GPUJobStore and check for errors.
	st, err := object.NewEventStore([]string{})

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create Event store: "+err.Error(),
		)

		return
	}

	// Ensure the GPUJobStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close Event store: %v\n", closeErr)
		}
	}()

	// List all GPUJobs in etcd.
	funcs, err := st.ListEvents(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list Events from etcd: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, funcs)
}

func CreateEvent(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var fc object.Event
	if err := c.BindJSON(&fc); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	fmt.Println(fc)
	// Create a new GPUJobStore and check for errors.
	st, err := object.NewEventStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create EventStore: "+err.Error(),
		)

		return
	}

	// Ensure the GPUJobStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close EventStore: %v\n", closeErr)
		}
	}()

	reply, err := st.GetEvents(
		c.Request.Context(),
		fc.Metadata.Namespace,
		fc.Metadata.Name,
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get Event from etcd: "+err.Error(),
		)

		return
	}

	if reply != nil {
		c.JSON(
			http.StatusConflict,
			"Create Event from file failed: same Job namespace & name",
		)

		return
	}

	if err := st.AddEvent(c.Request.Context(), &fc); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add Event to etcd: "+err.Error(),
		)

		return
	}
}

//nolint:dupl
func DeleteEvent(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var f object.Event
	if err := c.BindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create ReplicasetStore and check for errors.
	st, err := object.NewEventStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create Event store: "+err.Error(),
		)

		return
	}

	// Ensure the ReplicasetStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close Event store: %v\n", closeErr)
		}
	}()

	// Delete the replicaset from etcd.
	if err := st.DeleteEvent(
		c.Request.Context(),
		f.Metadata.Namespace,
		f.Metadata.Name,
	); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete Event from etcd: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Event deleted: "+f.Metadata.Name)
}
