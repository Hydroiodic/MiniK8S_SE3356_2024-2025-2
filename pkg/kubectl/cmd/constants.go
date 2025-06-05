package cmd

var kindStruct struct {
	Kind string `yaml:"kind"`
}

const (
	PodKind                     = "Pod"
	ServiceKind                 = "Service"
	ReplicaSetKind              = "ReplicaSet"
	DNSKind                     = "DNS"
	HorizontalPodAutoscalerKind = "HorizontalPodAutoscaler"
	GPUJobKind                  = "GpuJob"
	PVKind                      = "PersistentVolume"
	PVCKind                     = "PersistentVolumeClaim"
)

const (
	PodCmdName        = "pod"
	ServiceCmdName    = "service"
	ReplicaSetCmdName = "replicaset"
	DNSCmdName        = "dns"
	HPACmdName        = "hpa"
	PVCmdName         = "pv"
	PVCCmdName        = "pvc"
)

var (
	podHeaders = []string{
		"NAME",      // Pod.Metadata.Name
		"NAMESPACE", // Pod.Metadata.Namespace
		"STATUS",    // Pod.Status.Phase
		"AGE",       // Pod.Status.StartTime
		"IP",        // Pod.Status.IP
		"LABELS",    // Pod.Metadata.Labels
	}
	serviceHeaders = []string{
		"NAME",      // Service.Metadata.Name
		"NAMESPACE", // Service.Metadata.Namespace
		"STATUS",
		"CLUSTER-IP", // Service.Status.ClusterIP
		"SELECTOR",   // Service.Metadata.Labels
		"LABELS",     // Service.Metadata.Labels
	}
	dnsHeaders = []string{
		"NAME",      // DNS.Metadata.Name
		"NAMESPACE", // DNS.Metadata.Namespace
		"DOMAIN",    // DNS.Spec.Host
		"PATHS",     // DNS.Spec.Paths
		"LABELS",    // DNS.Metadata.Labels
	}
	replicaSetHeaders = []string{
		"NAME",       // ReplicaSet.Metadata.Name
		"NAMESPACE",  // ReplicaSet.Metadata.Namespace
		"DESIRED",    // ReplicaSet.Spec.Replicas
		"CURRENT",    // ReplicaSet.Status.Replicas
		"READY",      // ReplicaSet.Status.AvailableReplicas
		"CONTAINERS", // ReplicaSet.Spec.Template.Spec.Containers
		"IMAGES",     // ReplicaSet.Spec.Template.Spec.Containers
		"SELECTOR",   // ReplicaSet.Spec.Selector
		"LABELS",     // ReplicaSet.Metadata.Labels
	}
	hpaHeaders = []string{
		"NAME",      // HorizontalPodAutoscaler.Metadata.Name
		"NAMESPACE", // HorizontalPodAutoscaler.Metadata.Namespace
		"REFERENCE",
		/* HorizontalPodAutoscaler.Spec.ScaleTargetRef.Kind *
		 * HorizontalPodAutoscaler.Spec.ScaleTargetRef.Name */
		"TARGETS", // HorizontalPodAutoscaler.Spec.Metrics
		"MINPODS", // HorizontalPodAutoscaler.Spec.MinReplicas
		"MAXPODS", // HorizontalPodAutoscaler.Spec.MaxReplicas
		"LABELS",  // HorizontalPodAutoscaler.Metadata.Labels
	}
	pvHeaders = []string{
		"NAME",     // PersistentVolume.Metadata.Name
		"CAPACITY", // PersistentVolume.Spec.Capacity
		"TYPE",     // PersistentVolume.Spec.NFS
		"PATH",     // PersistentVolume.Spec.HostPath / PersistentVolume.Spec.NFS
		"STATUS",   // PersistentVolume.Status
	}
	pvcHeaders = []string{
		"NAME",        // PersistentVolumeClaim.Metadata.Name
		"NAMESPACE",   // PersistentVolumeClaim.Metadata.Namespace
		"STORAGE",     // PersistentVolumeClaim.Spec.Resources.Storage
		"VOLUME_NAME", // PersistentVolumeClaim.Spec.VolumeName
	}
)
