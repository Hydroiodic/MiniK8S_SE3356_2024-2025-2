package pod

import "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"

func combineRunAsUser(podUser, containerUser string) string {
	// Use the container user if it is set, otherwise use the pod user.
	if containerUser != "" {
		return containerUser
	}

	return podUser
}

func combineRunAsGroup(podGroup, containerGroup string) string {
	// Use the container group if it is set, otherwise use the pod group.
	if containerGroup != "" {
		return containerGroup
	}

	return podGroup
}

func combineFsGroup(podFsGroup, containerFsGroup string) string {
	// Use the container fsGroup if it is set, otherwise use the pod fsGroup.
	if containerFsGroup != "" {
		return containerFsGroup
	}

	return podFsGroup
}

func combineSupplementalGroups(
	podSupplementalGroups, containerSupplementalGroups []string,
) []string {
	// TODO: this function should not be configured in the container spec.
	if len(containerSupplementalGroups) > 0 {
		return containerSupplementalGroups
	}

	return podSupplementalGroups
}

func combineSecurityContexts(
	podSecurityContext, containerSecurityContext object.SecurityContext,
) object.SecurityContext {
	return object.SecurityContext{
		RunAsUser: combineRunAsUser(
			podSecurityContext.RunAsUser,
			containerSecurityContext.RunAsUser,
		),
		RunAsGroup: combineRunAsGroup(
			podSecurityContext.RunAsGroup,
			containerSecurityContext.RunAsGroup,
		),
		FsGroup: combineFsGroup(
			podSecurityContext.FsGroup,
			containerSecurityContext.FsGroup,
		),
		SupplementalGroups: combineSupplementalGroups(
			podSecurityContext.SupplementalGroups,
			containerSecurityContext.SupplementalGroups,
		),
	}
}
