package cmd

import (
	"bytes"
	"fmt"
	openai "github.com/fanook/aicli/internal/openapi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"strings"
	"text/template"
)

// genCmd 定义了 gen-cmd 命令
var genCmd = &cobra.Command{
	Use:     "gen-cmd [description]",
	Short:   "根据描述生成命令行指令及其解释",
	Long:    `根据用户提供的描述，使用 AI 生成适合当前机器的命令行指令，并提供相应的解释。用户可以查看后手动复制并执行这些命令。`,
	Example: `  acl gen-cmd "查看当前目录下的文件大小总和"`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		description := strings.Join(args, " ")

		apiKey := os.Getenv("AICLI_OPENAI_API_KEY")
		if apiKey == "" {
			logrus.Fatal("未设置 AICLI_OPENAI_API_KEY 环境变量")
		}

		templateStr, err := cmd.Flags().GetString("prompt")
		if err != nil {
			logrus.Fatalf("获取 template 标志失败: %v", err)
		}

		if templateStr == "" {
			templateStr = os.Getenv("AICLI_GENCMD_PROMPT")
		}

		if templateStr == "" {
			templateStr = "你是一个帮助生成命令行指令和解释的助手, 请根据以下描述生成一个适合当前机器的命令行指令，并提供简要的解释：描述：{{.Description}} 操作系统：{{.OS}} 架构：{{.Arch}}  生成的格式举例(严格按照此格式)： CMD: free -m \n 解释: 显示当前系统内存使用情况"
		}

		tmpl, err := template.New("gencmd").Parse(templateStr)
		if err != nil {
			logrus.Fatalf("解析模板失败: %v", err)
		}

		var promptBuffer bytes.Buffer
		err = tmpl.Execute(&promptBuffer, struct {
			Description string
			OS          string
			Arch        string
		}{
			Description: description,
			OS:          runtime.GOOS,
			Arch:        runtime.GOARCH,
		})
		if err != nil {
			logrus.Fatalf("执行模板失败: %v", err)
		}

		prompt := promptBuffer.String()

		response, err := openai.GenerateContent(apiKey, prompt)
		if err != nil {
			logrus.Fatalf("生成命令失败: %v", err)
		}

		fmt.Printf("\n%s\n\n", response)
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().StringP("prompt", "t", "", "自定义提示信息，例如: --prompt \"你的提示信息\"")
}
