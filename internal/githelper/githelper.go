package githelper

import (
	"os"
	"os/exec"
	"strings"
)

// GetGitChanges 获取当前 Git 仓库的变更信息
func GetGitChanges() (string, error) {
	cmd := exec.Command("git", "status", "--short")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GitCommit 执行 git commit 命令
func GitCommit(message string) error {
	// 确保所有变更被添加
	addCmd := exec.Command("git", "add", ".")
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	if err := addCmd.Run(); err != nil {
		return err
	}

	// 执行 commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	return commitCmd.Run()
}
