package cmd

import (
	"bytes"
	"fmt"
	"github.com/fanook/aicli/internal/provider"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"text/template"
)

var jokeCmd = &cobra.Command{
	Use:     "joke",
	Short:   "讲一个与程序员相关的笑话",
	Long:    `使用 AI 生成并展示一个与程序员相关的笑话。`,
	Example: `  acl joke`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("AICLI_OPENAI_API_KEY")
		if apiKey == "" {
			logrus.Fatal("未设置 AICLI_OPENAI_API_KEY 环境变量")
		}

		templateStr, err := cmd.Flags().GetString("prompt")
		if err != nil {
			logrus.Fatalf("获取 prompt 标志失败: %v", err)
		}

		if templateStr == "" {
			templateStr = os.Getenv("AICLI_JOKE_PROMPT")
		}

		if templateStr == "" {
			templateStr = "你是一个讲程序员相关笑话的助手, 请生成一个与程序员相关的笑话： 生成的格式举例（严格按照此格式）： [为什么程序员总是混淆圣诞节和万圣节？因为 Oct 31 == Dec 25！ 因为在八进制中，31 等于十进制的 25。]"
		}

		tmpl, err := template.New("joke").Parse(templateStr)
		if err != nil {
			logrus.Fatalf("解析模板失败: %v", err)
		}

		data := struct{}{}

		var promptBuffer bytes.Buffer
		err = tmpl.Execute(&promptBuffer, data)
		if err != nil {
			logrus.Fatalf("执行模板失败: %v", err)
		}

		prompt := promptBuffer.String()

		joke, err := provider.GenerateContent(prompt)
		if err != nil {
			logrus.Fatalf("生成笑话失败: %v", err)
		}

		fmt.Printf("😊%s😊\n", joke)
	},
}

func init() {
	rootCmd.AddCommand(jokeCmd)
	jokeCmd.Flags().StringP("prompt", "t", "", "自定义生成笑话的提示信息，例如: --prompt \"请生成一个有趣的程序员笑话：\"")
}
