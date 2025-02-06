package provider

import (
	"github.com/fanook/aicli/internal/deepseek"
	"github.com/fanook/aicli/internal/openai"
	"github.com/sirupsen/logrus"
	"os"
)

func GenerateContent(prompt string) (string, error) {
	provider := os.Getenv("AICLI_PROVIDER")
	if provider == "" {
		provider = "openai"
	}

	switch provider {
	case "openai":
		return openai.GenerateContent(prompt)
	case "deepseek":
		return deepseek.GenerateContent(prompt)
	default:
		logrus.Fatalf("未支持的 AI 提供商: %s, 请检查配置。", provider)
		return "", nil
	}
}
