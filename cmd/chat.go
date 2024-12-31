package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	openapi "github.com/fanook/aicli/internal/openapi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Conversation struct {
	History []string
}

var chatCmd = &cobra.Command{
	Use:     "chat",
	Short:   "与 AI 进行持续对话",
	Long:    `使用 AI 进行持续的对话，维持上下文和历史记录。`,
	Example: `  acl chat`,
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
			templateStr = os.Getenv("AICLI_CHAT_PROMPT")
		}

		if templateStr == "" {
			templateStr = "你是一个智能聊天助手，能够与用户进行自然流畅的对话。"
		}

		tmpl, err := template.New("chat").Parse(templateStr)
		if err != nil {
			logrus.Fatalf("解析模板失败: %v", err)
		}

		conversation := Conversation{
			History: []string{},
		}

		var promptBuffer bytes.Buffer
		err = tmpl.Execute(&promptBuffer, conversation)
		if err != nil {
			logrus.Fatalf("执行模板失败: %v", err)
		}

		initialPrompt := promptBuffer.String()
		conversation.History = append(conversation.History, fmt.Sprintf("系统: %s", initialPrompt))

		fmt.Println("😊 欢迎使用 AI 聊天助手！输入 'exit' 或 'quit' 退出对话。 😊")

		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("你: ")
			userInput, err := reader.ReadString('\n')
			if err != nil {
				logrus.Fatalf("读取输入失败: %v", err)
			}

			userInput = strings.TrimSpace(userInput)

			if userInput == "exit" || userInput == "quit" {
				fmt.Println("😊 再见！期待下次聊天。 😊")
				break
			}

			conversation.History = append(conversation.History, fmt.Sprintf("用户: %s", userInput))

			fullPrompt := strings.Join(conversation.History, "\n") + "\nAI:"

			reply, err := openapi.GenerateContent(apiKey, fullPrompt)
			if err != nil {
				logrus.Fatalf("生成回复失败: %v", err)
			}

			conversation.History = append(conversation.History, fmt.Sprintf("AI: %s", reply))

			fmt.Printf("🤖: %s\n", reply)
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.Flags().StringP("prompt", "t", "", "自定义初始化对话的提示信息，例如: --prompt \"你是一个友好的 AI 助手，能够帮助用户解决各种问题。\"")
}
