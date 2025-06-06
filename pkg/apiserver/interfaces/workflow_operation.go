package interfaces

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func GetWorkflows(c *gin.Context) {
	// Create GPUJobStore and check for errors.
	st, err := object.NewWorkflowStore([]string{})

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create Workflow store: "+err.Error(),
		)

		return
	}

	// Ensure the GPUJobStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close Workflow store: %v\n", closeErr)
		}
	}()

	// List all GPUJobs in etcd.
	funcs, err := st.ListWorkflows(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list Workflows from etcd: "+err.Error(),
		)

		return
	}

	// Convert the GPUJobs to a JSON format and return them.
	c.JSON(http.StatusOK, funcs)
}

func CreateWorkflow(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var fc object.Workflow
	if err := c.BindJSON(&fc); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	fmt.Println("createWorkflow")
	// Create a new GPUJobStore and check for errors.
	st, err := object.NewWorkflowStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create WorkflowStore: "+err.Error(),
		)

		return
	}

	// Ensure the GPUJobStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close WorkflowStore: %v\n", closeErr)
		}
	}()

	reply, err := st.GetWorkflows(
		c.Request.Context(),
		fc.Metadata.Namespace,
		fc.Metadata.Name,
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get Workflow from etcd: "+err.Error(),
		)

		return
	}

	if reply != nil {
		c.JSON(
			http.StatusConflict,
			"Create Workflow from file failed: same Job namespace & name",
		)

		return
	}

	if err := st.AddWorkflow(c.Request.Context(), &fc); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add Workflow to etcd: "+err.Error(),
		)

		return
	}
}

func DeleteWorkflow(c *gin.Context) {
	// Parse the JSON body into a Pod object.
	var f object.Workflow
	if err := c.BindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	fmt.Println(f)
	// Create ReplicasetStore and check for errors.
	st, err := object.NewWorkflowStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create Workflow store: "+err.Error(),
		)

		return
	}

	// Ensure the ReplicasetStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close Workflow store: %v\n", closeErr)
		}
	}()

	// Delete the replicaset from etcd.
	if err := st.DeleteWorkflow(
		c.Request.Context(),
		f.Metadata.Namespace,
		f.Metadata.Name,
	); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete Workflow from etcd: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Workflow deleted: "+f.Metadata.Name)
}
