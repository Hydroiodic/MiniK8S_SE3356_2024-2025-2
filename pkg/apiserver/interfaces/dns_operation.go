package interfaces

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

// Some local cache for the DNS and proxy information.
var (
	forwardingInfoLocalCache object.ForwardingInfo
	forwardingInfoNeedUpdate = true
	forwardingInfoMutex      sync.Mutex
)

func getProxyIPAddress() string {
	// We need to use the cluster IP of the proxy service.
	return object.ProxyClusterIP
}

func AddDNS(c *gin.Context) {
	// Create a new etcd connection for DNS operations.
	st, err := object.NewDNSStore([]string{})
	if err != nil {
		// Failed to create DNS store, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create DNS store: "+err.Error(),
		)

		return
	}

	// Ensure the DNS store is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close DNS store: %v\n", closeErr)
		}
	}()

	var dns object.DNS
	if err := c.ShouldBindJSON(&dns); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Combine the DNS record to the store.
	if err := st.CombineDNS(c.Request.Context(), &dns); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add DNS: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "DNS added successfully")
}

func DeleteDNS(c *gin.Context) {
	// Create a new etcd connection for DNS operations.
	st, err := object.NewDNSStore([]string{})
	if err != nil {
		// Failed to create DNS store, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create DNS store: "+err.Error(),
		)

		return
	}

	// Ensure the DNS store is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close DNS store: %v\n", closeErr)
		}
	}()

	// Bind the host parameter from the URL.
	var dns object.DNS
	if err := c.ShouldBindJSON(&dns); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := st.DeleteDNS(c.Request.Context(), dns.Spec.Host); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete DNS: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "DNS deleted successfully")
}

func GetForwardingInfo(c *gin.Context) {
	// Call `GetForwardingInfo` to retrieve the DNS resolve information.
	forwardingInfo, err := internalGetForwardingInfo()
	if err != nil {
		// Failed to get forwarding info, report error.
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get forwarding information: "+err.Error(),
		)

		return
	}

	// Send the forwarding information as a JSON response.
	c.JSON(http.StatusOK, forwardingInfo)
}

func internalGetForwardingInfo() (*object.ForwardingInfo, error) {
	// Lock the mutex to ensure thread safety.
	forwardingInfoMutex.Lock()
	defer forwardingInfoMutex.Unlock()

	// If the local cache is valid, return it directly.
	if !forwardingInfoNeedUpdate {
		return &forwardingInfoLocalCache, nil
	}

	// Create a context used for the etcd connection.
	ctx := context.Background()

	// Create a new DNS store for status checking.
	st, err := object.NewDNSStore([]string{})
	if err != nil {
		// Failed to create DNS store, report error.
		return nil, err
	}

	// Ensure the DNS store is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close DNS store: %v\n", closeErr)
		}
	}()

	// Create a new Service store for status checking.
	serviceStore, err := object.NewServiceStore([]string{})
	if err != nil {
		// Failed to create service store, report error.
		return nil, err
	}

	// Ensure the service store is closed after use.
	defer func() {
		if closeErr := serviceStore.Close(); closeErr != nil {
			log.Printf("Failed to close service store: %v\n", closeErr)
		}
	}()

	// Get all DNS objects from etcd.
	dnses, err := st.ListDNS(ctx)
	if err != nil {
		// Failed to list DNS, report error.
		return nil, err
	}

	// Get all valid services from etcd.
	services, err := serviceStore.ListServices(ctx, true)
	if err != nil {
		// Failed to list services, report error.
		return nil, err
	}

	// Make some lists for later use.
	dnsInfo := make([]object.DNSResolveInfo, 0)
	proxyInfo := make([]object.ProxyRule, 0)

	// Iterate through all DNS objects and find if the service exists.
	for _, dns := range dnses {
		// Check if the service exists.
		serviceExists := false

		// Iterate through all paths in the DNS object.
		for _, path := range dns.Spec.Paths {
			// Get the service with the same namespace and name.
			for _, service := range services {
				if service.Metadata.Namespace == dns.Metadata.Namespace &&
					service.Metadata.Name == path.ServiceName {
					// The service exists, we find a rule now.
					proxyInfo = append(proxyInfo, object.ProxyRule{
						Domain:     dns.Spec.Host,
						PathPrefix: path.Path,
						Target: service.Status.ClusterIP +
							":" + strconv.Itoa(path.ServicePort),
					})
					// Update the boolean flag.
					serviceExists = true

					break
				}
			}
		}

		if serviceExists {
			// As long as one path matches, we consider the DNS is valid.
			dnsInfo = append(dnsInfo, object.DNSResolveInfo{
				Host: dns.Spec.Host,
				IP:   getProxyIPAddress(),
			})
		}
	}

	// Update the local cache with the new forwarding information.
	forwardingInfoLocalCache = object.ForwardingInfo{
		DNSInfo:    dnsInfo,
		ProxyRules: proxyInfo,
	}
	// Mark the local cache as valid.
	forwardingInfoNeedUpdate = false

	return &forwardingInfoLocalCache, nil
}
