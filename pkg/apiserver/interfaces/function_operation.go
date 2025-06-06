package interfaces

import (
	"log"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func GetFunctions(c *gin.Context) {
	// Create GPUJobStore and check for errors.
	st, err := object.NewFuncStore([]string{})

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create Function store: "+err.Error(),
		)

		return
	}

	// Ensure the GPUJobStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close Function store: %v\n", closeErr)
		}
	}()

	// List all GPUJobs in etcd.
	funcs, err := st.ListFunctions(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list Functions from etcd: "+err.Error(),
		)

		return
	}

	// Convert the GPUJobs to a JSON format and return them.
	c.JSON(http.StatusOK, funcs)
}

//nolint:dupl
func CreateFunction(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var fc object.Function
	if err := c.BindJSON(&fc); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create a new GPUJobStore and check for errors.
	st, err := object.NewFuncStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create FuncStore: "+err.Error(),
		)

		return
	}

	// Ensure the GPUJobStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close FuncStore: %v\n", closeErr)
		}
	}()

	reply, err := st.GetFuntions(
		c.Request.Context(),
		fc.Metadata.Namespace,
		fc.Metadata.Name,
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get Function from etcd: "+err.Error(),
		)

		return
	}

	if reply != nil {
		c.JSON(
			http.StatusConflict,
			"Create Job from file failed: same Job namespace & name",
		)

		return
	}

	if err := st.AddFunction(c.Request.Context(), &fc); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add Function to etcd: "+err.Error(),
		)

		return
	}
}

//nolint:dupl
func DeleteFunction(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var f object.Function
	if err := c.BindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create ReplicasetStore and check for errors.
	st, err := object.NewFuncStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create function store: "+err.Error(),
		)

		return
	}

	// Ensure the ReplicasetStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close function store: %v\n", closeErr)
		}
	}()

	// Delete the replicaset from etcd.
	if err := st.DeleteFunction(
		c.Request.Context(),
		f.Metadata.Namespace,
		f.Metadata.Name,
	); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete function from etcd: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "function deleted: "+f.Metadata.Name)
}
