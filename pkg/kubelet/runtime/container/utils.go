package container

func combineUserAndGroup(user, group string) string {
	// This function combines user and group into a single string.
	separator := ":"
	if group == "" {
		separator = ""
	}

	return user + separator + group
}
