// internal/openapi/openai.go
package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// OpenAIRequest 定义发送给 OpenAI 的请求结构
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}

// OpenAIMessage 定义消息结构
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse 定义从 OpenAI 接收的响应结构
type OpenAIResponse struct {
	Choices []struct {
		Message OpenAIMessage `json:"message"`
	} `json:"choices"`
}

// getOpenAIConfig 读取 OpenAI 配置信息
func getOpenAIConfig() (string, string, error) {
	apiURL := os.Getenv("AICLI_OPENAI_API_URL")
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	model := os.Getenv("AICLI_OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}

	return apiURL, model, nil
}

// GenerateCommitMessage 使用 OpenAI 生成 commit 信息
func GenerateCommitMessage(apiKey, content string, prompt string) (string, error) {
	// 获取 OpenAI 配置
	apiURL, model, err := getOpenAIConfig()
	if err != nil {
		return "", err
	}

	requestBody := OpenAIRequest{
		Model: model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: content,
			},
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

	if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API 返回错误: %s", string(bodyBytes))
	}

	var openAIResp OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&openAIResp)
	if err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI API 返回空结果")
	}

	message := openAIResp.Choices[0].Message.Content
	message = strings.TrimSpace(message)
	return message, nil
}

// GenerateContent 使用 OpenAI 生成内容
func GenerateContent(apiKey, prompt string) (string, error) {
	apiURL, model, err := getOpenAIConfig()
	if err != nil {
		return "", err
	}

	requestBody := OpenAIRequest{
		Model: model,
		Messages: []OpenAIMessage{
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

	if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API 返回错误: %s", string(bodyBytes))
	}

	var openAIResp OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&openAIResp)
	if err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI API 返回空结果")
	}

	message := openAIResp.Choices[0].Message.Content
	message = strings.TrimSpace(message)
	return message, nil
}
