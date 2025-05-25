package interfaces

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"path"
	"strconv"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/etcd"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func randomGenClusterIP() string {
	cip := "222.111." + strconv.Itoa(
		rand.Intn(256),
	) + "." + strconv.Itoa(
		rand.Intn(256),
	)

	return cip
}

func genClusterIP() (string, error) {
	// Create a new context for the request.
	ctx := context.Background()

	// Create a new ClusterIP
	clusterIPStore, err := object.NewClusterIPStore([]string{})
	if err != nil {
		return "", err
	}

	// Ensure the ClusterIPStore is closed after use.
	defer func() {
		if closeErr := clusterIPStore.Close(); closeErr != nil {
			log.Printf("Failed to close cluster IP store: %v", closeErr)
		}
	}()

	var clusterIP string
	// Repeatedly generate a random ClusterIP until it is not already in use.
	for {
		// Generate a random ClusterIP.
		clusterIP = randomGenClusterIP()

		// Check if the ClusterIP is already in use.
		res, err := clusterIPStore.GetClusterIP(ctx, clusterIP)
		if err != nil {
			log.Printf("Failed to get cluster IP: %v", err)
			continue
		}

		// If the ClusterIP is already in use, continue to the next iteration.
		if res != "" {
			log.Printf("Cluster IP %s is already in use, continue", clusterIP)
			continue
		}

		// Store the generated ClusterIP in etcd.
		err = clusterIPStore.SetClusterIP(ctx, clusterIP)
		if err != nil {
			log.Printf("Failed to set cluster IP: %v", err)
			continue
		}

		// Jump out of the loop if no error occurred.
		break
	}

	return clusterIP, nil
}

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

	// Generate a new ClusterIP for the service.
	clusterIP, err := genClusterIP()
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
