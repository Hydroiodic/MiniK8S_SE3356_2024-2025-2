package cmd

import (
	"fmt"
	"strings"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver/interfaces"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an instance of a Kubernetes resource",
	Run: func(_ *cobra.Command, args []string) {
		// Check if the correct number of arguments is provided.
		if len(args) < 2 {
			fmt.Println("Usage: kubectl delete <type> <name> [namespace]")
			return
		}

		// Get the resource type and name from the arguments.
		instanceType := args[0]
		instanceName := args[1]
		instanceNamespace := interfaces.DefaultNamespace

		// If a namespace is provided, use it.
		if len(args) >= 3 {
			instanceNamespace = args[2]
		}

		switch strings.ToLower(instanceType) {
		case PodCmdName:
			deletePod(instanceName, instanceNamespace)
		case ServiceCmdName:
			deleteService(instanceName, instanceNamespace)
		case ReplicaSetCmdName:
			deleteReplicaSet(instanceName)
		case DNSCmdName:
			deleteDNS(instanceName, instanceNamespace)
		case HPACmdName:
			deleteHPA(instanceName)
		case PVCmdName:
			deletePV(instanceName)
		case PVCCmdName:
			deletePVC(instanceName, instanceNamespace)
		default:
			fmt.Printf("Unknown resource type: %s\n", instanceType)
			return
		}
	},
}

// Register the delete command with the root command.
func init() {
	rootCmd.AddCommand(deleteCmd)
}

// Delete Pod.
func deletePod(name, namespace string) {
	err := apiserver.NewAPIClient("").DeletePodByName(name, namespace)
	if err != nil {
		fmt.Println("Error deleting pod:", err)
	}
}

// Delete Service.
func deleteService(name, namespace string) {
	err := apiserver.NewAPIClient("").DeleteServiceByName(name, namespace)
	if err != nil {
		fmt.Println("Error deleting service:", err)
	}
}

// Delete ReplicaSet.
func deleteReplicaSet(name string) {
	err := apiserver.NewAPIClient("").DeleteReplicaset(name)
	if err != nil {
		fmt.Println("Error deleting replicaset:", err)
	}
}

// Delete DNS configuration.
func deleteDNS(name, namespace string) {
	err := apiserver.NewAPIClient("").DeleteDNSByName(name, namespace)
	if err != nil {
		fmt.Println("Error deleting DNS:", err)
		return
	}
}

// Delete HPA configuration.
func deleteHPA(name string) {
	// TODO: Implement HPA deletion logic.
	fmt.Printf("WIP: Deleting HPA %s\n", name)
}

// Delete PersistentVolume.
func deletePV(name string) {
	err := apiserver.NewAPIClient("").DeletePVByName(name)
	if err != nil {
		fmt.Println("Error deleting persistent volume:", err)
		return
	}

	fmt.Printf("PersistentVolume %s deleted successfully.\n", name)
}

// Delete PersistentVolumeClaim.
func deletePVC(name, namespace string) {
	err := apiserver.NewAPIClient("").
		DeletePVCByName(namespace, name)
	if err != nil {
		fmt.Println("Error deleting persistent volume claim:", err)
		return
	}

	fmt.Printf(
		"PersistentVolumeClaim %s in namespace %s deleted successfully.\n",
		name,
		namespace,
	)
}
