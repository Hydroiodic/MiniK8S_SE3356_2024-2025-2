package cmd

import (
	"fmt"
	"strings"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Display one or more Kubernetes instances",
	Run: func(_ *cobra.Command, args []string) {
		// Check for the correct number of arguments.
		if len(args) < 1 {
			fmt.Println("Usage: kubectl get <type>")
			return
		}

		// Initial value for Kubectl parameters.
		resourceType := args[0]

		// Now handle the resource type and name.
		switch strings.ToLower(resourceType) {
		case NodeCmdName, NodeCmdName + "s":
			getAllNodes()
		case PodCmdName, PodCmdName + "s":
			getAllPods()
		case ServiceCmdName, ServiceCmdName + "s":
			getAllServices()
		case ReplicaSetCmdName, ReplicaSetCmdName + "s":
			getAllReplicaSets()
		case DNSCmdName:
			getAllDNS()
		case HPACmdName, HPACmdName + "s":
			getAllHPA()
		case PVCmdName, PVCmdName + "s":
			getAllPV()
		case PVCCmdName, PVCCmdName + "s":
			getALLPVC()
		case FunctionCmdName + "s":
			getAllFunctions()
		case JobCmdName + "s":
			getJobs()
		default:
			fmt.Printf("Unknown resource type: %s\n", resourceType)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func getJobs() {
	// Get all Nodes from the API server.
	results, err := apiserver.NewAPIClient("").GetAllGPUJob()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, j := range results {
		fmt.Printf("jobname : %s , status : %s\n", j.Metadata.Name, j.Status)
	}
}

func getAllFunctions() {
	// Get all Nodes from the API server.
	results, err := apiserver.NewAPIClient("").GetAllFunction()
	if err != nil {
		fmt.Println(err)
		return
	}

	printFuncs(results)
}
func getAllNodes() {
	// Get all Nodes from the API server.
	results, err := apiserver.NewAPIClient("").GetNodes()
	if err != nil {
		fmt.Println(err)
		return
	}

	printNodes(results)
}

func getAllPods() {
	// Get all Pods from the API server.
	results, err := apiserver.NewAPIClient("").GetPods()
	if err != nil {
		fmt.Println(err)
		return
	}

	printPods(results)
}

func getAllServices() {
	// Get all Services from the API server.
	results, err := apiserver.NewAPIClient("").GetServices()
	if err != nil {
		fmt.Println(err)
		return
	}

	printServices(results)
}

func getAllDNS() {
	// Get all DNS records from the API server.
	dnsRecords, err := apiserver.NewAPIClient("").GetDNS()
	if err != nil {
		fmt.Println(err)
		return
	}

	printDNS(dnsRecords)
}

func getAllReplicaSets() {
	// Get all ReplicaSets from the API server.
	rs, err := apiserver.NewAPIClient("").GetReplicasets()
	if err != nil {
		fmt.Println(err)
		return
	}

	printReplicaSets(rs)
}

func getAllHPA() {
	// Get all Horizontal Pod Autoscalers (HPAs) from the API server.
	hpas, err := apiserver.NewAPIClient("").GetHpas()
	if err != nil {
		fmt.Println(err)
		return
	}

	printHPAs(hpas)
}

func getAllPV() {
	// Get all Persistent Volumes (PVs) from the API server.
	pvs, err := apiserver.NewAPIClient("").GetPVs()
	if err != nil {
		fmt.Println(err)
		return
	}

	printPVs(pvs)
}

func getALLPVC() {
	// Get all Persistent Volume Claims (PVCs) from the API server.
	pvcs, err := apiserver.NewAPIClient("").GetPVCs()
	if err != nil {
		fmt.Println(err)
		return
	}

	printPVCs(pvcs)
}
