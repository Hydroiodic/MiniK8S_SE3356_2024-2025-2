// Package cmd implements basic functionality for MiniK8S.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "Kubectl CLI tool for MiniK8S",
}

// Execute runs the root command of the kubectl CLI.
func Execute() {
	// The CLI arguments are passed to the root command.
	rootCmd.SetArgs(os.Args[1:])

	// Execute the root command.
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
