package cmd

import (
	"fmt"
	"os"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var execCmd = &cobra.Command{
	Use:   "apply",
	Short: "Interactive exec into a resource",
	Run: func(cmd *cobra.Command, args []string) {

		fileFlag, _ := cmd.Flags().GetString("file") // 获取 `-f` 参数的值
		if fileFlag != "" {
			fmt.Println("Using file:", fileFlag)
			pod_config, err := parse_yaml(fileFlag)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				pass_to_api_server(pod_config)
			}
		} else {
			fmt.Println("No file provided.")
		}
		// **手动重置 flag**
		cmd.Flags().Lookup("file").Value.Set("")
	},
}

func pass_to_api_server(podconfig *object.Pod) {

}

// 在 `init()` 里添加 `-f` flag
func init() {
	execCmd.Flags().StringP("file", "f", "", "Specify the configuration file")
	rootCmd.AddCommand(execCmd)
}

func parse_yaml(file_addr string) (*object.Pod, error) {
	data, err := os.ReadFile(file_addr)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	var podConfig object.Pod

	// 解析 YAML
	err = yaml.Unmarshal(data, &podConfig)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return nil, err
	}

	// 输出解析结果
	fmt.Printf("Pod Name: %s\n", podConfig.Metadata.Name)
	fmt.Printf("Namespace: %s\n", podConfig.Metadata.Namespace)

	fmt.Println("Containers:")
	for _, container := range podConfig.Spec.Containers {
		fmt.Printf("- Name: %s, Image: %s\n", container.Name, container.Image)
		if len(container.Ports) > 0 {
			fmt.Printf("  Ports: %d\n", container.Ports[0].ContainerPort)
		}
	}
	return &podConfig, err
}
