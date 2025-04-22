package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [resource]",
	Short: "Display resources (e.g., pods, nodes)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resource := args[0]
		fmt.Printf("Listing all %s...\n", resource)
		// 模拟数据
		fmt.Println("- pod-1")
		fmt.Println("- pod-2")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
