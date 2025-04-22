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

func Execute() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Interactive mode started. Type 'exit' to quit.")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			fmt.Println("Exiting interactive mode.")
			break
		}

		args := strings.Fields(input) // 按空格拆分成参数列表
		rootCmd.SetArgs(args)         // 让 `rootCmd` 解析这些参数
		err := rootCmd.Execute()      // 执行 `cobra` 解析

		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}
