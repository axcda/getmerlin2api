package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestChatCompletion(t *testing.T) {
	url := "http://localhost:8080/v1/chat/completions"

	// 准备请求数据
	requestBody := map[string]interface{}{
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "你好，请做个自我介绍",
			},
		},
		"stream": false,
		"model":  "gpt-3.5-turbo",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// 发送请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// 检查响应内容
	if response["choices"] == nil {
		t.Error("Response does not contain 'choices' field")
	}

	// 打印响应内容以供查看
	t.Logf("Response: %+v", response)
}
