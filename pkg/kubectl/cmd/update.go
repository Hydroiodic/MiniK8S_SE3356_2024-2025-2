package cmd

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "update an exist function",
	Example: "kubectl update -f config.yaml",
	Run: func(cmd *cobra.Command, _ []string) {

		var f object.Function

		updateContainer(f)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func updateContainer(f object.Function) {
	// ci := apiserver.NewAPIClient("")
	// pods, err := ci.GetAllGPUJob()

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// for _,pod
}
