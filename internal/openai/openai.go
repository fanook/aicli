package openai

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

// getOpenAIConfig 读取 OpenAI 配置信息
func getOpenAIConfig() (string, string, string, error) {
	apiURL := os.Getenv("AICLI_OPENAI_API_URL")
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	model := os.Getenv("AICLI_OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}

	apiKey := os.Getenv("AICLI_OPENAI_API_KEY")
	if apiKey == "" {
		logrus.Fatal("未设置 AICLI_OPENAI_API_KEY 环境变量")
	}

	return apiURL, apiKey, model, nil
}

func GenerateContent(prompt string) (string, error) {
	apiURL, apiKey, model, err := getOpenAIConfig()
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

	if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API 返回错误: %s", string(bodyBytes))
	}

	var openAIResp Response
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
