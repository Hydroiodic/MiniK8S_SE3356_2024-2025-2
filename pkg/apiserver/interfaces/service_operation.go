package interfaces

import (
	"context"
	"log"
	"net/http"
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func removeClusterIP(clusterIP string) error {
	// Create a new context for the request.
	ctx := context.Background()

	// Create a new ClusterIPStore instance.
	clusterIPStore, err := object.NewClusterIPStore([]string{})
	if err != nil {
		return err
	}

	// Ensure the ClusterIPStore is closed after use.
	defer func() {
		if closeErr := clusterIPStore.Close(); closeErr != nil {
			log.Printf("Failed to close cluster IP store: %v", closeErr)
		}
	}()

	// Remove the ClusterIP from etcd.
	err = clusterIPStore.DeleteClusterIP(ctx, clusterIP)
	if err != nil {
		return err
	}

	return nil
}

func CreateService(c *gin.Context) {
	// Bind the JSON request body to the Service struct.
	var svc object.Service
	if err := c.BindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Set the default namespace if not provided.
	if svc.Metadata.Namespace == "" {
		svc.Metadata.Namespace = DefaultNamespace
	}

	// Create a new ServiceStore instance.
	st, _ := object.NewServiceStore([]string{})

	// Ensure the ServiceStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			c.JSON(
				http.StatusInternalServerError,
				"Failed to close service store: "+closeErr.Error(),
			)
		}
	}()

	// Check if the service already exists.
	reply, err := st.GetServiceWithoutStatus(
		c.Request.Context(),
		svc.Metadata.Namespace,
		svc.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get service: "+err.Error(),
		)

		return
	}

	if reply != nil {
		c.JSON(
			http.StatusConflict,
			"Service already exists: "+svc.Metadata.Name,
		)

		return
	}

	// Check if the labels of the service are valid.
	// NOTE: It should not contain `internal` because it is reserved.
	if _, ok := svc.Metadata.Labels[object.SERVICE_INTERNEL_LABEL]; ok {
		c.JSON(
			http.StatusBadRequest,
			"Invalid label: 'internal' is reserved and cannot be used.",
		)

		return
	}

	// Generate a new ClusterIP for the service.
	clusterIP, err := object.GenClusterIP()
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to generate cluster IP: "+err.Error(),
		)

		return
	}

	svc.Status.ClusterIP = clusterIP

	// Apply the service to etcd
	err = st.AddService(c.Request.Context(), &svc, false)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add service: "+err.Error(),
		)

		return
	}

	path := path.Join(etcd.ClusterIPPrefix, clusterIP)
	c.JSON(http.StatusOK, "Service created: "+path)
}

func DeleteService(c *gin.Context) {
	// Bind the JSON request body to the Service struct.
	var svc object.Service
	if err := c.BindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Set the default namespace if not provided.
	if svc.Metadata.Namespace == "" {
		svc.Metadata.Namespace = DefaultNamespace
	}

	// Create a new ServiceStore instance.
	st, err := object.NewServiceStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create service store: "+err.Error(),
		)

		return
	}

	// Ensure the ServiceStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			c.JSON(
				http.StatusInternalServerError,
				"Failed to close service store: "+closeErr.Error(),
			)
		}
	}()

	// Check if the service exists.
	reply, err := st.GetServiceWithoutStatus(
		c.Request.Context(),
		svc.Metadata.Namespace,
		svc.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get service: "+err.Error(),
		)

		return
	}

	// If the service does not exist, return a 404 error.
	if reply == nil {
		c.JSON(
			http.StatusNotFound,
			"Service not found: "+svc.Metadata.Name,
		)

		return
	}

	err = st.DeleteServiceWithoutStatus(
		c.Request.Context(),
		svc.Metadata.Namespace,
		svc.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete service: "+err.Error(),
		)

		return
	}

	// Release the ClusterIP.
	err = removeClusterIP(svc.Status.ClusterIP)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to remove cluster IP: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "Service deleted: "+svc.Metadata.Name)
}

func GetAllService(c *gin.Context) {
	// Create a new ServiceStore instance.
	st, err := object.NewServiceStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create service store: "+err.Error(),
		)

		return
	}

	// Ensure the ServiceStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			c.JSON(
				http.StatusInternalServerError,
				"Failed to close service store: "+closeErr.Error(),
			)
		}
	}()

	// Get all services from etcd.
	svcs, err := st.ListServicesWithoutStatus(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get all services: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, svcs)
}
