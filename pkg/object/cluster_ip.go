package object

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
)

const (
	// ClusterIPPrefix is the prefix for ClusterIP addresses.
	ClusterIPPrefix = "222.111."
	// We should reserve two IP addresses for DNS and Proxy.
	DNSClusterIP   = ClusterIPPrefix + "1.1"
	ProxyClusterIP = ClusterIPPrefix + "1.2"
	// The maximum number of ClusterIPs that can be generated.
	MaxClusterIPs = 256 * 256 // 256 * 256 = 65536 possible ClusterIPs
	MaxTryTimes   = 1000      // Maximum attempts to generate a unique ClusterIP
)

type ClusterIP struct {
	ClusterIPMap map[string]any `json:"clusterIPMap"`
}

func randomClusterIP(existingClusterIPs map[string]any) (string, error) {
	// Check if the number of existing ClusterIPs is less than the maximum.
	if len(existingClusterIPs) >= MaxClusterIPs {
		log.Println(
			"Maximum number of ClusterIPs reached, cannot generate more.",
		)

		return "", fmt.Errorf(
			"maximum number of ClusterIPs reached: %d",
			MaxClusterIPs,
		)
	}

	// Only try `MaxTryTimes` times.
	i := 0

	// Repeatedly generate a random ClusterIP until it is not already in use.
	for {
		// Generate a random ClusterIP in the format "222.111.x.y".
		clusterIP := ClusterIPPrefix +
			strconv.Itoa(
				rand.Intn(256),
			) + "." +
			strconv.Itoa(
				rand.Intn(256),
			)

		// Check if the generated ClusterIP is already in use.
		if _, exists := existingClusterIPs[clusterIP]; !exists {
			// If not, return the generated ClusterIP.
			return clusterIP, nil
		}

		// If it is already in use, continue to generate a new one.
		log.Printf(
			"ClusterIP %s is already in use, generating a new one...",
			clusterIP,
		)

		// If `MaxTryTimes` is reached, return an error.
		if i := i + 1; i > MaxTryTimes {
			return "", fmt.Errorf("max try times %d reached", MaxTryTimes)
		}
	}
}

// TODO: IP may be wasted if interruption occurs during generation.
func GenClusterIP() (string, error) {
	// Create a new context for the request.
	ctx := context.Background()

	// Create a new ClusterIP
	clusterIPStore, err := NewClusterIPStore([]string{})
	if err != nil {
		return "", err
	}

	// Ensure the ClusterIPStore is closed after use.
	defer func() {
		if closeErr := clusterIPStore.Close(); closeErr != nil {
			log.Printf("Failed to close cluster IP store: %v", closeErr)
		}
	}()

	// Get all existing ClusterIPs from etcd.
	existingClusterIPs, err := clusterIPStore.GetClusterIPStruct(ctx)
	if err != nil {
		log.Printf("Failed to list existing cluster IPs: %v", err)
		return "", err
	}

	// If there are no existing ClusterIPs, we should create it.
	if existingClusterIPs == nil {
		existingClusterIPs = &ClusterIP{
			ClusterIPMap: make(map[string]any),
		}
	}

	// Generate a new ClusterIP that is not already in use.
	newClusterIP, err := randomClusterIP(existingClusterIPs.ClusterIPMap)
	if err != nil {
		log.Printf("Failed to generate a new ClusterIP: %v", err)
		return "", err
	}

	// Add the new ClusterIP to the store.
	err = clusterIPStore.SetClusterIP(ctx, newClusterIP)
	if err != nil {
		log.Printf("Failed to add new ClusterIP %s: %v", newClusterIP, err)
		return "", err
	}

	return newClusterIP, nil
}
