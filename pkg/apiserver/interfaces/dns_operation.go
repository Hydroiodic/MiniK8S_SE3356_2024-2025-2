package interfaces

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func getProxyIPAddress() string {
	// If one domain has multiple service paths,
	// we need to redirect to our HTTP proxy server.
	return object.ProxyClusterIP
}

func resolveDNS(dns *object.DNS) (string, error) {
	// If the domain has more than one path, we need to use the Nginx IP address.
	if len(dns.Spec.Paths) == 1 {
		return dns.Spec.Paths[0].ServiceIP, nil
	} else if len(dns.Spec.Paths) > 1 {
		return getProxyIPAddress(), nil
	}

	// If the domain has no path, we need to return an error.
	return "", fmt.Errorf("no path found for domain %s", dns.Spec.Host)
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

	// Get the DNS object from the store.
	dnsObj, err := st.GetDNS(c.Request.Context(), dns.Spec.Host)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get DNS: "+err.Error(),
		)

		return
	}

	// Get the IP address for the domain.
	ip, err := resolveDNS(dnsObj)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to parse DNS: "+err.Error(),
		)

		return
	}

	// Create a DNSResolveInfo object.
	dnsResolveInfo := object.DNSResolveInfo{
		Host: dnsObj.Spec.Host,
		IP:   ip,
	}

	// Marshal the DNSResolveInfo object to JSON.
	dnsResolveInfoJSON, err := json.Marshal(dnsResolveInfo)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to marshal DNSResolveInfo: "+err.Error(),
		)

		return
	}

	// Send a message in RabbitMQ to notify DNS Server.
	err = mqtemplate.SendMessageToQueue(
		mqtemplate.UpdateHostQueueName,
		string(dnsResolveInfoJSON),
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to send message to RabbitMQ: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "DNS added successfully")
}

func DeleteSingleDNS(c *gin.Context) {
	// Parse the JSON body into a DNS object.
	var dns object.DNS
	if err := c.ShouldBindJSON(&dns); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check there is one and only one path in the DNS object.
	if len(dns.Spec.Paths) != 1 {
		c.JSON(
			http.StatusBadRequest,
			"Invalid DNS object: expected one path, got "+fmt.Sprint(
				len(dns.Spec.Paths),
			),
		)

		return
	}

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

	// Delete the DNS record from the store.
	err = st.DeleteSinglePathDNS(
		c.Request.Context(),
		dns.Spec.Host,
		dns.Spec.Paths[0].Path,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to delete DNS: "+err.Error(),
		)

		return
	}

	// Get the DNS object from the etcd store.
	dnsObj, err := st.GetDNS(c.Request.Context(), dns.Spec.Host)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get DNS: "+err.Error(),
		)

		return
	}

	// NOTE: If the DNS object is nil, `IP` is set to "0",
	//       which means the DNS object is deleted.
	dnsResolveInfo := object.DNSResolveInfo{
		Host: dns.Spec.Host,
		IP:   "0",
	}

	// Resolve the DNS object to get the IP address.
	if dnsObj != nil {
		// Get the IP address for the domain.
		ip, err := resolveDNS(dnsObj)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				"Failed to parse DNS: "+err.Error(),
			)

			return
		}

		dnsResolveInfo.IP = ip
	}

	// Marshal the DNSResolveInfo object to JSON.
	dnsResolveInfoJSON, err := json.Marshal(dnsResolveInfo)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to marshal DNSResolveInfo: "+err.Error(),
		)

		return
	}

	// Send a message in RabbitMQ to notify DNS Server.
	err = mqtemplate.SendMessageToQueue(
		mqtemplate.UpdateHostQueueName,
		string(dnsResolveInfoJSON),
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to send message to RabbitMQ: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "DNS deleted successfully")
}

func GetDNSResolve(c *gin.Context) {
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

	// Get the DNS object from the store.
	dns, err := st.ListDNS(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get DNS: "+err.Error(),
		)

		return
	}

	// Create a slice to store the DNS resolve information.
	dnsResolveInfoList := make([]object.DNSResolveInfo, 0, len(dns))

	// Iterate over the DNS records and extract the host and IP address.
	for _, record := range dns {
		// Convert the domain name to lowercase.
		domain := strings.ToLower(record.Spec.Host)

		// Resolve the DNS record to get the IP address.
		result, err := resolveDNS(record)
		if err != nil {
			log.Printf(
				"Failed to resolve DNS for domain %s: %v\n",
				domain,
				err,
			)

			continue
		}

		// Add the resolved DNS information to the list.
		dnsResolveInfoList = append(
			dnsResolveInfoList,
			object.DNSResolveInfo{
				Host: domain,
				IP:   result,
			},
		)
	}

	// Send the DNS resolve information as a JSON response.
	c.JSON(http.StatusOK, dnsResolveInfoList)
}
