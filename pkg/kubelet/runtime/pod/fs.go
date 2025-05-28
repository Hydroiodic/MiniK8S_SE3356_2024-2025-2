package pod

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func chownRecursive(path string, uid, gid int) error {
	return filepath.Walk(
		path,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Change the ownership of the file or directory.
			return os.Chown(p, uid, gid)
		},
	)
}

func chmodDirRecursive(path string, mode os.FileMode) error {
	return filepath.Walk(
		path,
		func(p string, info os.FileInfo, err error) error {
			// If any error occurs while walking the path, return it.
			if err != nil {
				return err
			}
			// If the path is a directory, change its permissions.
			if info.IsDir() {
				return os.Chmod(p, mode)
			}
			// If the path is not a directory, do nothing.
			return nil
		},
	)
}

func createDirIfNotExist(dirPath string) error {
	// Check if the directory exists.
	_, err := os.Stat(dirPath)
	if err == nil {
		// Directory exists, return nil.
		fmt.Printf("Directory %s already exists.\n", dirPath)
		return nil
	}

	if !os.IsNotExist(err) {
		// An error occurred other than "not exists".
		fmt.Printf("Failed to stat directory %s: %v\n", dirPath, err)
		return err
	}

	// Create the directory.
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		fmt.Printf("Failed to create directory %s: %v\n", dirPath, err)
		return err
	}

	fmt.Printf("Directory %s created successfully.\n", dirPath)

	return nil
}

func directorySetgid(dirPath, gid string) error {
	// Check if the directory exists.
	info, err := os.Stat(dirPath)
	if err != nil {
		fmt.Printf("Failed to stat directory %s: %v\n", dirPath, err)
		return err
	}

	if gid != "" {
		// Convert the gid to an integer and change the group ownership of the directory.
		gidInt, convErr := strconv.Atoi(gid)
		if convErr != nil {
			fmt.Printf("Failed to convert gid %s to int: %v\n", gid, convErr)
			return convErr
		}

		// Change the group ownership of the directory recursively.
		err = chownRecursive(dirPath, -1, gidInt)
		if err != nil {
			fmt.Printf(
				"Failed to change group ownership of directory %s to %s: %v\n",
				dirPath,
				gid,
				err,
			)

			return err
		}
	}

	// Get the current permissions of the directory.
	currentMode := info.Mode()

	// Add the setgid bit to the current permissions.
	newMode := currentMode | os.ModeSetgid

	// Set the new permissions on the directory.
	err = chmodDirRecursive(dirPath, newMode)
	if err != nil {
		fmt.Printf(
			"Failed to set g+s permission on directory %s: %v\n",
			dirPath,
			err,
		)

		return err
	}

	return nil
}
