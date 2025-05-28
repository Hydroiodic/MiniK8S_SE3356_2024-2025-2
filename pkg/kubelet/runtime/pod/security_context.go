package pod

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/image"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func combineRunAsUser(podUser, containerUser string) string {
	// If both of them are empty, return "0" (root user).
	if podUser == "" && containerUser == "" {
		return "0"
	}

	// Use the container user if it is set, otherwise use the pod user.
	if containerUser != "" {
		return containerUser
	}

	return podUser
}

func combineRunAsGroup(podGroup, containerGroup string) string {
	// If both of them are empty, return "0" (root group).
	if podGroup == "" && containerGroup == "" {
		return "0"
	}

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
	podSecurityContext, containerSecurityContext object.SecurityContext,
) []string {
	// NOTE: the `supplementalGroups` option should not be set in container.
	// At first let's combine their `fsGroup`.
	fsGroup := combineFsGroup(
		podSecurityContext.FsGroup,
		containerSecurityContext.FsGroup,
	)

	// The result should not be nil.
	result := []string{}

	// If `podSecurityContext.SupplementalGroups` isn't nil, append it.
	if podSecurityContext.SupplementalGroups != nil {
		result = append(result, podSecurityContext.SupplementalGroups...)
	}

	// If `fsGroup` isn't empty, append it.
	if fsGroup != "" {
		result = append(result, fsGroup)
	}

	return result
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
			podSecurityContext,
			containerSecurityContext,
		),
	}
}

func processSecurityContextsMerge(
	podSecurityContext, containerSecurityContext object.SecurityContext,
	imgRef string,
) (*object.SecurityContext, error) {
	// Merge the security contexts.
	// If there're any previously defined groups, we should use them.
	// NOTE: If we do not provide the primary group, the container will act as expected.
	combinedSecurityContext := combineSecurityContexts(
		podSecurityContext,
		containerSecurityContext,
	)

	// If the field `RunAsGroup` is not set, there's nothing else to do.
	if combinedSecurityContext.RunAsGroup == "" {
		// NOTE: actually, this would never happen now.
		return &combinedSecurityContext, nil
	}

	// The `RunAsGroup` field is set, so we need to merge the previously defined groups.
	username, _, err := image.GetCurrentUserGroup(
		combinedSecurityContext.RunAsUser,
		imgRef,
	)
	if err != nil {
		return nil, err
	}

	// If the user is not found, there's nothing else to do.
	if username == "" {
		return &combinedSecurityContext, nil
	}

	// If the user is found, we should get the supplemental groups for it.
	supplementalGroups, err := image.GetCurrentUserSupplymentalGroups(
		username,
		imgRef,
	)
	if err != nil {
		return nil, err
	}

	// Append the supplemental groups to the security context.
	combinedSecurityContext.SupplementalGroups = append(
		combinedSecurityContext.SupplementalGroups,
		supplementalGroups...,
	)

	return &combinedSecurityContext, nil
}

func processSecurityContextsStrict(
	podSecurityContext, containerSecurityContext object.SecurityContext,
	imgRef string,
) (*object.SecurityContext, error) {
	// Strictly set the security contexts.
	// If there're any previously defined groups, we should abort them.
	combinedSecurityContext := combineSecurityContexts(
		podSecurityContext,
		containerSecurityContext,
	)

	// If the field `RunAsGroup` is set, there's nothing else to do.
	if combinedSecurityContext.RunAsGroup != "" {
		return &combinedSecurityContext, nil
	}

	// The `RunAsGroup` field is not set, we need to add one to enforce no other groups.
	_, gid, err := image.GetCurrentUserGroup(
		combinedSecurityContext.RunAsUser,
		imgRef,
	)
	if err != nil {
		return nil, err
	}

	// Set the `RunAsGroup` field to the gid of the user.
	// If the user is not found, we should set it to "0" (root group).
	combinedSecurityContext.RunAsGroup = "0"
	if gid != "" {
		combinedSecurityContext.RunAsGroup = gid
	}

	return &combinedSecurityContext, nil
}

func processSecurityContexts(
	podSecurityContext, containerSecurityContext object.SecurityContext,
	imgRef string,
) (*object.SecurityContext, error) {
	// If the `SupplementalGroupsPolicy` is not set, we should use the default one.
	supplymentalGroupsPolicy := containerSecurityContext.SupplementalGroupsPolicy
	if supplymentalGroupsPolicy == "" {
		supplymentalGroupsPolicy = podSecurityContext.SupplementalGroupsPolicy
	}

	// Process the security contexts based on the policy.
	switch supplymentalGroupsPolicy {
	case object.SupplementalGroupsStrict:
		return processSecurityContextsStrict(
			podSecurityContext,
			containerSecurityContext,
			imgRef,
		)
	default:
		return processSecurityContextsMerge(
			podSecurityContext,
			containerSecurityContext,
			imgRef,
		)
	}
}
