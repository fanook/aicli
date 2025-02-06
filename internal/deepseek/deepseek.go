// internal/deepseek/deepseek.go
package deepseek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func getDeepseekConfig() (string, string, string, error) {
	apiURL := os.Getenv("AICLI_DEEPSEEK_API_URL")
	if apiURL == "" {
		apiURL = "https://api.deepseek.com/chat/completions"
	}

	model := os.Getenv("AICLI_DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}

	apiKey := os.Getenv("AICLI_DEEPSEEK_API_KEY")
	if apiKey == "" {
		logrus.Fatal("未设置 AICLI_DEEPSEEK_API_KEY 环境变量")
	}

	return apiURL, apiKey, model, nil
}

func GenerateContent(prompt string) (string, error) {
	apiURL, apiKey, model, err := getDeepseekConfig()
	if err != nil {
		return "", err
	}

	requestBody := Request{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("Deepseek API 返回错误: %s", string(bodyBytes))
	}

	var deepseekResp Response
	err = json.NewDecoder(resp.Body).Decode(&deepseekResp)
	if err != nil {
		return "", err
	}

	if len(deepseekResp.Choices) == 0 {
		return "", fmt.Errorf("Deepseek API 返回空结果")
	}

	message := deepseekResp.Choices[0].Message.Content
	message = strings.TrimSpace(message)
	return message, nil
}
