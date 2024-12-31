package cmd

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd 是应用的根命令
var rootCmd = &cobra.Command{
	Use:   "acl",
	Short: "ACL 是一个集合多种 AI 工具的命令行应用",
	Long:  `ACL 提供了多个基于 AI 的工具，帮助开发者提高工作效率。`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	err := godotenv.Load()
	if err != nil {

	}
}
