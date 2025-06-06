package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/jedib0t/go-pretty/v6/table"
)

func stringSliceToInterface(slice []string) []interface{} {
	// Convert a slice of strings to a slice of empty interfaces.
	result := make([]any, len(slice))
	for i, v := range slice {
		result[i] = v
	}

	return result
}

func printTable(headers []string, rows [][]string) {
	// Create a new table writer.
	t := table.NewWriter()
	// Set the output to standard output.
	t.SetOutputMirror(os.Stdout)

	// Append the headers to the table.
	t.AppendHeader(table.Row(stringSliceToInterface(headers)))

	// Convert each row to a table.Row and append it to the table.
	for _, row := range rows {
		// If the row is empty, print a separator row.
		if len(row) == 0 {
			t.AppendSeparator()
			continue
		}

		t.AppendRow(table.Row(stringSliceToInterface(row)))
	}

	// Render the table to standard output.
	t.SetStyle(table.StyleLight)
	t.Render()
}

func convertStringMapToString(m map[string]string) string {
	// Convert the map to a string representation.
	keys := make([]string, 0, len(m))
	for k, v := range m {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}

	// Sort the keys of the map to ensure a consistent order.
	sort.Strings(keys)

	// Join the keys with "\n" to create a multi-line string.
	return strings.Join(keys, "\n")
}

func printNodes(nodes []object.Kubelet) {
	// Sort the nodes by their names.
	sort.SliceStable(nodes, func(i, j int) bool {
		// Compare the names of the nodes.
		return nodes[i].Config.Name < nodes[j].Config.Name
	})

	// Prepare the rows for the table.
	rows := [][]string{}
	// Iterate over the nodes and prepare the rows for the table.
	for i := range nodes {
		rows = append(rows, []string{
			nodes[i].Config.Name,
			"READY", // Node status is always "Ready"
			nodes[i].StartTime.Format(time.RFC3339),
			time.Since(nodes[i].StartTime).Truncate(time.Second).String(),
			nodes[i].LastUpdateTime.Format(time.RFC3339),
		})

		// Add a separator row between nodes.
		if i != len(nodes)-1 {
			rows = append(rows, []string{})
		}
	}

	// Print the table with headers.
	printTable(nodeHeaders, rows)
}

func printPods(pods []object.Pod) {
	// Sort the pods by their start time.
	sort.SliceStable(pods, func(i, j int) bool {
		// Compare the start times of the pods.
		return pods[i].Status.StartTime.After(pods[j].Status.StartTime)
	})

	// Use a stable sort to sort pods by their start time.
	rows := [][]string{}
	// Iterate over the pods and prepare the rows for the table.
	for i, pod := range pods {
		rows = append(rows, []string{
			pod.Metadata.Name,
			pod.Metadata.Namespace,
			pod.Status.Phase,
			time.Since(pod.Status.StartTime).Truncate(time.Second).String(),
			pod.Status.IP,
			convertStringMapToString(pod.Metadata.Labels),
		})

		// Add a separator row between pods.
		if i != len(pods)-1 {
			rows = append(rows, []string{})
		}
	}

	// Print the table with headers.
	printTable(podHeaders, rows)
}

func printServices(services []object.Service) {
	// Sort the services by their names.
	sort.SliceStable(services, func(i, j int) bool {
		// Compare the names of the services.
		return services[i].Metadata.Name < services[j].Metadata.Name
	})

	// Prepare the rows for the table.
	rows := [][]string{}
	// Iterate over the services and prepare the rows for the table.
	for i, svc := range services {
		// Get the service status.
		status := "Pending"
		if svc.CheckReady() {
			status = "Active"
		}

		rows = append(rows, []string{
			svc.Metadata.Name,
			svc.Metadata.Namespace,
			svc.Type,
			status,
			svc.Status.ClusterIP,
			strings.Join(svc.GetEndpoints(), ","),
			convertStringMapToString(svc.Spec.Selector),
			convertStringMapToString(svc.Metadata.Labels),
			strings.Join(svc.GetPorts(), ","),
			strings.Join(svc.GetTargetPorts(), ","),
			strings.Join(svc.GetNodePorts(), ","),
		})

		// Add a separator row between services.
		if i != len(services)-1 {
			rows = append(rows, []string{})
		}
	}

	// Print the table with headers.
	printTable(serviceHeaders, rows)
}

func printDNS(dns []object.DNS) {
	// Sort the DNS records by their host names.
	sort.SliceStable(dns, func(i, j int) bool {
		// Compare the host names of the DNS records.
		return dns[i].Spec.Host < dns[j].Spec.Host
	})

	// Prepare the rows for the table.
	rows := [][]string{}
	// Iterate over the DNS records and prepare the rows for the table.
	for i, d := range dns {
		// Get the paths for the DNS record.
		paths := []string{}
		for _, path := range d.Spec.Paths {
			paths = append(
				paths,
				fmt.Sprintf(
					"%s -> %s:%d",
					path.Path,
					path.ServiceName,
					path.ServicePort,
				),
			)
		}

		rows = append(rows, []string{
			d.Metadata.Name,
			d.Metadata.Namespace,
			d.Spec.Host,
			strings.Join(paths, "\n"),
			convertStringMapToString(d.Metadata.Labels),
		})

		// Add a separator row between DNS records.
		if i != len(dns)-1 {
			rows = append(rows, []string{})
		}
	}

	// Print the table with headers.
	printTable(dnsHeaders, rows)
}

func printReplicaSets(replicaSets []object.ReplicaSet) {
	// Sort the ReplicaSets by their names.
	sort.SliceStable(replicaSets, func(i, j int) bool {
		// Compare the names of the ReplicaSets.
		return replicaSets[i].Metadata.Name < replicaSets[j].Metadata.Name
	})

	// Prepare the rows for the table.
	rows := [][]string{}
	// Iterate over the ReplicaSets and prepare the rows for the table.
	for i, rs := range replicaSets {
		// Get container names and images.
		containerNames := []string{}
		containerImages := []string{}
		// Iterate over the containers in the ReplicaSet's Pod template.
		for _, c := range rs.Spec.Template.Spec.Containers {
			containerNames = append(containerNames, c.Name)
			containerImages = append(containerImages, c.Image)
		}

		// TODO: `rs.Status.AvailableReplicas` are shown twice here?
		rows = append(rows, []string{
			rs.Metadata.Name,
			rs.Metadata.Namespace,
			strconv.Itoa(rs.Spec.Replicas),
			strconv.Itoa(rs.Status.AvailableReplicas),
			strconv.Itoa(rs.Status.AvailableReplicas),
			strings.Join(containerNames, "\n"),
			strings.Join(containerImages, "\n"),
			convertStringMapToString(rs.Spec.Selector),
			convertStringMapToString(rs.Metadata.Labels),
		})

		// Add a separator row between ReplicaSets.
		if i != len(replicaSets)-1 {
			rows = append(rows, []string{})
		}
	}

	// Print the table with headers.
	printTable(replicaSetHeaders, rows)
}

func printHPAs(hpas []object.HorizontalPodAutoscaler) {
	// Sort the HPAs by their names.
	sort.SliceStable(hpas, func(i, j int) bool {
		// Compare the names of the HPAs.
		return hpas[i].Metadata.Name < hpas[j].Metadata.Name
	})

	// Prepare the rows for the table.
	rows := [][]string{}
	// Iterate over the HPAs and prepare the rows for the table.
	for i, hpa := range hpas {
		// Get the targets for the HPA.
		targets := []string{}
		for _, metric := range hpa.Spec.Metrics {
			targets = append(
				targets,
				fmt.Sprintf(
					"%s:%f%%",
					metric.Resource.Name,
					*metric.Resource.Target.AverageUtilization,
				),
			)
		}

		rows = append(rows, []string{
			hpa.Metadata.Name,
			hpa.Metadata.Namespace,
			fmt.Sprintf(
				"%s/%s",
				hpa.Spec.ScaleTargetRef.Kind,
				hpa.Spec.ScaleTargetRef.Name,
			),
			strings.Join(targets, "\n"),
			strconv.Itoa(int(hpa.Spec.MinReplicas)),
			strconv.Itoa(int(hpa.Spec.MaxReplicas)),
			convertStringMapToString(hpa.Metadata.Labels),
		})

		// Add a separator row between HPAs.
		if i != len(hpas)-1 {
			rows = append(rows, []string{})
		}
	}

	// Print the table with headers.
	printTable(hpaHeaders, rows)
}

func printPVs(pvs []object.PersistentVolume) {
	// Sort the PersistentVolumes by their names.
	sort.SliceStable(pvs, func(i, j int) bool {
		return pvs[i].Metadata.Name < pvs[j].Metadata.Name
	})

	rows := [][]string{}

	for i, pv := range pvs {
		// 获取存储类型和路径
		storageType := ""
		storagePath := ""

		if pv.Spec.NFS != nil {
			storageType = "NFS"
			storagePath = fmt.Sprintf(
				"%s:%s",
				pv.Spec.NFS.Server,
				pv.Spec.NFS.Path,
			)
		} else if pv.Spec.HostPath != nil {
			storageType = "HostPath"
			storagePath = pv.Spec.HostPath.Path
		}

		rows = append(rows, []string{
			pv.Metadata.Name,
			pv.Spec.Capacity.Storage,
			storageType,
			storagePath,
			pv.Status,
		})

		if i != len(pvs)-1 {
			rows = append(rows, []string{})
		}
	}

	printTable(pvHeaders, rows)
}

func printPVCs(pvcs []object.PersistentVolumeClaim) {
	// Sort the PersistentVolumeClaims by their names.
	sort.SliceStable(pvcs, func(i, j int) bool {
		return pvcs[i].Metadata.Name < pvcs[j].Metadata.Name
	})

	rows := [][]string{}

	for i, pvc := range pvcs {
		// 获取绑定的 PV 名称
		volumeName := pvc.Spec.VolumeName
		if volumeName == "" {
			volumeName = "-"
		}

		rows = append(rows, []string{
			pvc.Metadata.Name,
			pvc.Metadata.Namespace,
			pvc.Spec.Capacity.Storage,
			volumeName,
		})

		if i != len(pvcs)-1 {
			rows = append(rows, []string{})
		}
	}

	printTable(pvcHeaders, rows)
}
