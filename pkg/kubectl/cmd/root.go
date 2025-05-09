// Package cmd implements basic functionality for MiniK8S.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "An interactive CLI like kubectl",
}
var kubectlCmd = &cobra.Command{
	Use:   "kubectl",
	Short: "Kubernetes command line tool",
}

// Execute runs the root command of the kubectl CLI.
func Execute() {
	reader := bufio.NewReader(os.Stdin)
	_, _ = fmt.Println("Interactive mode started. Type 'exit' to quit.")

	for {
		_, _ = fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			_, _ = fmt.Println("Exiting interactive mode.")
			break
		}

		args := strings.Fields(input) // 按空格拆分成参数列表
		rootCmd.SetArgs(args)         // 让 `rootCmd` 解析这些参数
		err := rootCmd.Execute()      // 执行 `cobra` 解析

		if err != nil {
			_, _ = fmt.Println("Error:", err)
		}
	}
}

func init() {
	// 将 kubectlCmd 添加到 rootCmd，使它成为所有命令的前缀
	rootCmd.AddCommand(kubectlCmd)
}
