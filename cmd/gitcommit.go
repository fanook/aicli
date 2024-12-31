package cmd

import (
	"bytes"
	"fmt"
	"github.com/fanook/aicli/internal/githelper"
	openapi "github.com/fanook/aicli/internal/openapi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// gcCmd 定义了 gc 命令
var gcCmd = &cobra.Command{
	Use:     "git-cmt",
	Short:   "使用Git Changes生成commit 信息",
	Long:    `根据当前 Git 仓库的变更，使用 AI 生成合适的 Git commit 信息，允许用户编辑后自动执行提交命令。`,
	Example: `  acl git-cmt`,
	Run: func(cmd *cobra.Command, args []string) {
		changes, err := githelper.GetGitChanges()
		if err != nil {
			logrus.Fatalf("获取 Git 变更信息失败: %v", err)
		}

		if changes == "" {
			logrus.Info("当前没有任何变更，无需提交。")
			return
		}

		logrus.Infof("当前变更:\n%s", changes)

		apiKey := os.Getenv("AICLI_OPENAI_API_KEY")
		if apiKey == "" {
			logrus.Fatal("未设置 AICLI_OPENAI_API_KEY 环境变量")
		}

		templateStr, err := cmd.Flags().GetString("prompt")
		if err != nil {
			logrus.Fatalf("获取 template 标志失败: %v", err)
		}

		if templateStr == "" {
			templateStr = os.Getenv("AICLI_GITCOMMIT_PROMPT")
		}

		if templateStr == "" {
			templateStr = "你是一个帮助生成 Git commit 信息的助手。请根据以下 Git 仓库的变更生成一个简洁且有意义的 Git commit 信息。请严格遵循以下格式，并且只能使用以下两种类别：\n\n[类别] 描述\n\n**可用类别：**\n- **feat**: 新功能\n- **fix**: 修复\n\n**示例：**\n[fix] 修复用户登录时的验证错误\n[feat] 添加用户个人资料页面\n\n变更内容：\n{{.Changes}}"
		}

		tmpl, err := template.New("commit").Parse(templateStr)
		if err != nil {
			logrus.Fatalf("解析模板失败: %v", err)
		}

		var promptBuffer bytes.Buffer
		err = tmpl.Execute(&promptBuffer, struct {
			Changes string
		}{
			Changes: changes,
		})
		if err != nil {
			logrus.Fatalf("执行模板失败: %v", err)
		}

		prompt := promptBuffer.String()

		commitMessage, err := openapi.GenerateContent(apiKey, prompt)
		if err != nil {
			logrus.Fatalf("生成 commit 信息失败: %v", err)
		}

		fmt.Printf("\n生成的 commit 信息:\n%s\n\n", commitMessage)

		tmpFile, err := ioutil.TempFile("", "commit_message_*.txt")
		if err != nil {
			logrus.Fatalf("创建临时文件失败: %v", err)
		}
		defer os.Remove(tmpFile.Name()) // 程序结束时删除临时文件

		// 写入初始 commit 信息
		_, err = tmpFile.WriteString(commitMessage)
		if err != nil {
			logrus.Fatalf("写入临时文件失败: %v", err)
		}
		tmpFile.Close()

		editor := os.Getenv("EDITOR")
		if editor == "" {
			// 如果未设置 EDITOR 环境变量，使用默认的编辑器
			if isCommandAvailable("vim") {
				editor = "vim"
			} else if isCommandAvailable("nano") {
				editor = "nano"
			} else {
				logrus.Fatal("未设置 EDITOR 环境变量，且系统中未安装 vim 或 nano。")
			}
		}

		cmdEditor := exec.Command(editor, tmpFile.Name())
		cmdEditor.Stdin = os.Stdin
		cmdEditor.Stdout = os.Stdout
		cmdEditor.Stderr = os.Stderr

		if err := cmdEditor.Run(); err != nil {
			logrus.Fatalf("打开编辑器失败: %v", err)
		}

		finalCommitBytes, err := ioutil.ReadFile(tmpFile.Name())
		if err != nil {
			logrus.Fatalf("读取临时文件失败: %v", err)
		}

		finalCommitMessage := strings.TrimSpace(string(finalCommitBytes))

		if finalCommitMessage == "" {
			logrus.Fatal("Commit 信息为空，取消提交。")
		}

		fmt.Printf("\n最终的 commit 信息:\n%s\n\n", finalCommitMessage)

		// 提示用户确认
		fmt.Println("按下回车键以确认并执行 Git 提交，或按其他键取消。")
		fmt.Print("确认提交？[Y/n]: ")

		// 捕捉用户输入
		var userInput string
		fmt.Scanln(&userInput)

		userInput = strings.ToLower(strings.TrimSpace(userInput))

		if userInput == "n" || userInput == "no" {
			logrus.Info("操作已取消。")
			return
		}

		addCmd := exec.Command("git", "add", ".")
		addOutput, err := addCmd.CombinedOutput()
		if err != nil {
			logrus.Fatalf("执行 'git add .' 失败: %v\n输出: %s", err, string(addOutput))
		}
		logrus.Info("'git add .' 执行成功。")

		commitCmd := exec.Command("git", "commit", "-m", finalCommitMessage)
		commitOutput, err := commitCmd.CombinedOutput()
		if err != nil {
			logrus.Fatalf("执行 'git commit' 失败: %v\n输出: %s", err, string(commitOutput))
		}
		logrus.Info("'git commit' 执行成功。")
	},
}

func init() {
	rootCmd.AddCommand(gcCmd)
	gcCmd.Flags().StringP("prompt", "t", "", "自定义生成commit的提示信息，例如: --prompt \"[fix] {{.Changes}}\"")
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
