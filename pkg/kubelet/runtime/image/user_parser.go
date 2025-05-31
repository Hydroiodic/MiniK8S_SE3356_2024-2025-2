package image

import (
	"slices"
	"strings"
)

// `GetCurrentUserGroup` retrieves the GID of the user specified by
// `usernameOrID` from the `/etc/passwd` file in the specified image.
// It returns a tuple containing the username and the GID as strings.
// If the user is not found, it returns an empty string for both.
func GetCurrentUserGroup(usernameOrID, imgRef string) (string, string, error) {
	// Let's get the contents of the /etc/passwd file from the image.
	content, err := ReadFileFromImage(imgRef, PASSWD_FILE_PATH)
	if err != nil {
		return "", "", err
	}

	// If the content is empty, return an empty slice.
	if content == "" {
		return "", "", nil
	}

	// Now let's parse the content.
	// The `passwd` file is expected to have lines in the format:
	//     username:x:uid:gid:comment:home:shell
	// We should return the GID as a string slice.
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// Trim whitespace.
		line = strings.TrimSpace(line)
		// Skip empty lines.
		if line == "" {
			continue
		}

		// Split the line by colon and check if it has enough parts.
		parts := strings.Split(line, ":")
		// Skip malformed lines
		if len(parts) < 4 {
			continue
		}

		// The first part is the username,
		// the third part is the UID,
		// and the fourth part is the GID.
		username := parts[0]
		uid := parts[2] // UID is the third field
		gid := parts[3] // GID is the fourth field

		if username == usernameOrID || uid == usernameOrID {
			// If we found the user, return the GID.
			return username, gid, nil
		}
	}

	// If no user was found, return an empty string.
	return "", "", nil
}

func GetCurrentUserSupplementalGroups(
	username, imgRef string,
) ([]string, error) {
	// Let's get the contents of the /etc/group file from the image.
	content, err := ReadFileFromImage(imgRef, GROUP_FILE_PATH)
	if err != nil {
		return nil, err
	}

	// If the content is empty, return an empty slice.
	if content == "" {
		return nil, nil
	}

	// Now let's parse the content.
	// The `group` file is expected to have lines in the format:
	//     groupname:x:gid:user1,user2,...
	lines := strings.Split(content, "\n")

	var groups []string

	for _, line := range lines {
		// Trim whitespace.
		line = strings.TrimSpace(line)
		// Skip empty lines.
		if line == "" {
			continue
		}

		// Split the line by colon and check if it has enough parts.
		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			continue
		}

		gid := parts[2]                       // GID is the third field.
		users := strings.Split(parts[3], ",") // Users are in the fourth field.

		// Check if the username is in the list of users.
		if slices.Contains(users, username) {
			// GID instead of Group Name is ok.
			groups = append(groups, gid)
		}
	}

	return groups, nil
}
