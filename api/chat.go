package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rubleowen/GetMerlin2Api/auth"
)

var (
	tokenCache      string
	tokenExpiry     time.Time
	tokenMutex      sync.Mutex
	requestCache    = make(map[string]ImageGenerationResult)
	requestCacheMux sync.RWMutex
)

type OpenAIRequest struct {
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Delta struct {
	Content string `json:"content,omitempty"`
	Role    string `json:"role,omitempty"`
}

type Choice struct {
	Delta        Delta  `json:"delta"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason,omitempty"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type MerlinRequest struct {
	Attachments []interface{} `json:"attachments"`
	ChatID      string        `json:"chatId"`
	Language    string        `json:"language"`
	Message     struct {
		Content  string `json:"content"`
		Context  string `json:"context"`
		ChildID  string `json:"childId"`
		ID       string `json:"id"`
		ParentID string `json:"parentId"`
	} `json:"message"`
	Metadata struct {
		LargeContext  bool `json:"largeContext"`
		MerlinMagic   bool `json:"merlinMagic"`
		ProFinderMode bool `json:"proFinderMode"`
		WebAccess     bool `json:"webAccess"`
		Temperature   int  `json:"temperature"`
	} `json:"metadata"`
	Mode  string `json:"mode"`
	Model string `json:"model"`
}

type MerlinResponse struct {
	Data struct {
		Content string `json:"content"`
	} `json:"data"`
}

type OpenAIResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

type ImageGenerationRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

type ImageGenerationResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64_json,omitempty"`
	} `json:"data"`
}

type MerlinImageRequest struct {
	Action struct {
		Message struct {
			Attachments []interface{} `json:"attachments"`
			Content     string        `json:"content"`
			Metadata    struct {
				Context string `json:"context"`
			} `json:"metadata"`
			ParentId string `json:"parentId"`
			Role     string `json:"role"`
		} `json:"message"`
		Type string `json:"type"`
	} `json:"action"`
	ChatId   string `json:"chatId"`
	Mode     string `json:"mode"`
	Settings struct {
		MerlinPromptMagic bool `json:"merlinPromptMagic"`
		ModelConfig       []struct {
			AspectRatio    string `json:"aspectRatio"`
			ModelId        string `json:"modelId"`
			NumberOfImages int    `json:"numberOfImages"`
		} `json:"modelConfig"`
		NegativePrompt string `json:"negativePrompt"`
	} `json:"settings"`
}

type ImageGenerationResult struct {
	Response ImageGenerationResponse
	Err      error
}

type MerlinImageGenerationRequest struct {
	Action struct {
		Message struct {
			Attachments []interface{} `json:"attachments"`
			Content     string        `json:"content"`
			Metadata    struct {
				Context string `json:"context"`
			} `json:"metadata"`
			ParentId string `json:"parentId"`
			Role     string `json:"role"`
		} `json:"message"`
		Type string `json:"type"`
	} `json:"action"`
	ChatId   string `json:"chatId"`
	Mode     string `json:"mode"`
	Settings struct {
		MerlinPromptMagic bool `json:"merlinPromptMagic"`
		ModelConfig       []struct {
			AspectRatio    string `json:"aspectRatio"`
			ModelId        string `json:"modelId"`
			NumberOfImages int    `json:"numberOfImages"`
		} `json:"modelConfig"`
		NegativePrompt string `json:"negativePrompt"`
	} `json:"settings"`
}

// OpenAI 兼容的响应结构
type OpenAIImageResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64_json,omitempty"`
	} `json:"data"`
}

type OpenAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
		Param   string `json:"param,omitempty"`
	} `json:"error"`
}

type OpenAIStreamResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// OpenAI 图片生成请求结构
type OpenAIImageGenerationRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
	Model  string `json:"model,omitempty"`
}

// OpenAI 图片生成响应结构
type OpenAIImageGenerationResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url"`
		B64JSON       string `json:"b64_json"`
		RevisedPrompt string `json:"revised_prompt"`
	} `json:"data"`
}

type WallflowerRequest struct {
	Feature struct {
		ModelConfig struct {
			AspectRatio    string `json:"aspectRatio"`
			ModelId        string `json:"modelId"`
			NumberOfImages int    `json:"numberOfImages"`
		} `json:"modelConfig"`
		Type string `json:"type"`
	} `json:"feature"`
	IsPublic bool   `json:"isPublic"`
	Prompt   string `json:"prompt"`
	Style    string `json:"style"`
}

// 定义通用的URL结构体
type ImageURL struct {
	URL string `json:"url"`
}

// 定义通用的响应结构体
type ImageResponse struct {
	Created int64      `json:"created"`
	Data    []ImageURL `json:"data"`
}

type WallflowerResponse struct {
	Payload []struct {
		Variations []struct {
			URL  string `json:"url"`
			IID  string `json:"iid"`
			Seed int    `json:"seed"`
		} `json:"variations"`
		// 其他字段...
	} `json:"payload"`
}

func getTokenWithCache() (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	// 如果token还在有效期内,直接返回
	if tokenCache != "" && time.Now().Before(tokenExpiry) {
		return tokenCache, nil
	}

	// 获取新token
	token, err := auth.GenerateToken()
	if err != nil {
		return "", err
	}

	// 更新缓存
	tokenCache = token
	tokenExpiry = time.Now().Add(30 * time.Minute)

	return token, nil
}

func getCachedImageResult(prompt string) (ImageGenerationResult, bool) {
	requestCacheMux.RLock()
	defer requestCacheMux.RUnlock()

	result, exists := requestCache[prompt]
	return result, exists
}

func cacheImageResult(prompt string, result ImageGenerationResult) {
	requestCacheMux.Lock()
	defer requestCacheMux.Unlock()

	requestCache[prompt] = result
}

func generateUUID() string {
	return uuid.New().String()
}

func generateImage(w http.ResponseWriter, flusher http.Flusher, prompt string, model string) {
	log.Printf("开始生成图片，提示词: %s, 模型: %s", prompt, model)

	// 设置响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("x-request-id", generateUUID())
	log.Printf("响应头设置完成")

	// 创建带超时的HTTP客户端
	client := &http.Client{
		Timeout: 60 * time.Second, // 增加超时时间到60秒
		Transport: &http.Transport{
			TLSHandshakeTimeout: 10 * time.Second,
			IdleConnTimeout:     90 * time.Second,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}

	// 获取 session token
	sessionToken := os.Getenv("MERLIN_SESSION_TOKEN")
	if sessionToken == "" {
		log.Printf("错误: MERLIN_SESSION_TOKEN 未设置")
		sendErrorResponse(w, "MERLIN_SESSION_TOKEN is not set", "internal_error", http.StatusInternalServerError)
		return
	}
	log.Printf("成功获取 session token")

	// 构建 session 请求
	sessionReq, err := http.NewRequest("GET", "https://session.getmerlin.in/?from=web", nil)
	if err != nil {
		log.Printf("错误: 创建session请求失败: %v", err)
		sendErrorResponse(w, fmt.Sprintf("create session request failed: %v", err), "internal_error", http.StatusInternalServerError)
		return
	}
	log.Printf("session请求创建成功")

	// 设置 session 请求头
	sessionReq.Header.Set("accept", "application/json, text/plain, */*")
	sessionReq.Header.Set("accept-language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	sessionReq.Header.Set("cache-control", "no-cache")
	sessionReq.Header.Set("origin", "https://www.getmerlin.in")
	sessionReq.Header.Set("pragma", "no-cache")
	sessionReq.Header.Set("referer", "https://www.getmerlin.in/")
	sessionReq.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	sessionReq.Header.Set("sec-ch-ua-mobile", "?0")
	sessionReq.Header.Set("sec-ch-ua-platform", `"macOS"`)
	sessionReq.Header.Set("sec-fetch-dest", "empty")
	sessionReq.Header.Set("sec-fetch-mode", "cors")
	sessionReq.Header.Set("sec-fetch-site", "same-site")
	sessionReq.Header.Set("cookie", fmt.Sprintf("__Secure-authjs.session-token=%s", sessionToken))
	log.Printf("session请求头设置完成")

	// 发送 session 请求
	log.Printf("正在发送session请求...")
	sessionResp, err := client.Do(sessionReq)
	if err != nil {
		log.Printf("错误: session请求失败: %v", err)
		sendErrorResponse(w, fmt.Sprintf("session request failed: %v", err), "internal_error", http.StatusInternalServerError)
		return
	}
	defer sessionResp.Body.Close()
	log.Printf("session请求发送成功，状态码: %d", sessionResp.StatusCode)

	if sessionResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(sessionResp.Body)
		log.Printf("错误: session请求返回非200状态码: %d, 响应体: %s", sessionResp.StatusCode, string(body))
		sendErrorResponse(w, fmt.Sprintf("session request returned non-200 status code: %d", sessionResp.StatusCode), "internal_error", sessionResp.StatusCode)
		return
	}

	sessionBody, err := io.ReadAll(sessionResp.Body)
	if err != nil {
		log.Printf("错误: 读取session响应失败: %v", err)
		sendErrorResponse(w, fmt.Sprintf("read session response failed: %v", err), "internal_error", http.StatusInternalServerError)
		return
	}
	log.Printf("session响应读取成功: %s", string(sessionBody))

	var sessionData struct {
		User struct {
			AccessToken string `json:"accessToken"`
		} `json:"user"`
	}
	if err = json.Unmarshal(sessionBody, &sessionData); err != nil {
		log.Printf("错误: 解析session响应失败: %v", err)
		sendErrorResponse(w, fmt.Sprintf("decode session response failed: %v, response: %s", err, string(sessionBody)), "internal_error", http.StatusInternalServerError)
		return
	}

	if sessionData.User.AccessToken == "" {
		log.Printf("错误: 收到空的access token")
		sendErrorResponse(w, fmt.Sprintf("received empty access token, response: %s", string(sessionBody)), "internal_error", http.StatusInternalServerError)
		return
	}
	log.Printf("成功获取access token")

	// 根据模型名称选择对应的ModelId
	var modelId string
	switch model {
	case "recraft-v3":
		modelId = "fal-ai/recraft-v3"
	case "flux-1.1-pro":
		modelId = "black-forest-labs/flux-1.1-pro"
	default:
		// 如果没有指定模型，使用用户传入的模型名称
		modelId = model
	}

	log.Printf("使用模型: %s", modelId)

	// 构造新的 Wallflower 请求
	reqBody := WallflowerRequest{
		Feature: struct {
			ModelConfig struct {
				AspectRatio    string `json:"aspectRatio"`
				ModelId        string `json:"modelId"`
				NumberOfImages int    `json:"numberOfImages"`
			} `json:"modelConfig"`
			Type string `json:"type"`
		}{
			ModelConfig: struct {
				AspectRatio    string `json:"aspectRatio"`
				ModelId        string `json:"modelId"`
				NumberOfImages int    `json:"numberOfImages"`
			}{
				AspectRatio:    "1:1",
				ModelId:        modelId,
				NumberOfImages: 2,
			},
			Type: "GENERATE",
		},
		IsPublic: false,
		Prompt:   prompt,
		Style:    "Auto",
	}
	log.Printf("Wallflower请求体构造完成: %+v", reqBody)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("错误: 序列化请求体失败: %v", err)
		sendErrorResponse(w, fmt.Sprintf("error marshaling request body: %v", err), "internal_error", http.StatusInternalServerError)
		return
	}
	log.Printf("请求体序列化成功: %s", string(jsonData))

	httpReq, err := http.NewRequest("POST", "https://arcane.getmerlin.in/v1/wallflower/unified-generation", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("错误: 创建HTTP请求失败: %v", err)
		sendErrorResponse(w, fmt.Sprintf("error creating request: %v", err), "internal_error", http.StatusInternalServerError)
		return
	}
	log.Printf("HTTP请求创建成功")

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Origin", "https://www.getmerlin.in")
	httpReq.Header.Set("Pragma", "no-cache")
	httpReq.Header.Set("Priority", "u=1, i")
	httpReq.Header.Set("Referer", "https://www.getmerlin.in/")
	httpReq.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	httpReq.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	httpReq.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	httpReq.Header.Set("Sec-Fetch-Dest", "empty")
	httpReq.Header.Set("Sec-Fetch-Mode", "cors")
	httpReq.Header.Set("Sec-Fetch-Site", "same-site")
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sessionData.User.AccessToken))
	httpReq.Header.Set("x-merlin-version", "web-merlin")
	log.Printf("HTTP请求头设置完成")

	log.Printf("正在发送图片生成请求...")
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("错误: 发送请求失败: %v", err)
		sendErrorResponse(w, fmt.Sprintf("send request failed: %v", err), "internal_error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	log.Printf("请求发送成功，状态码: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("错误: 服务器返回非200状态码: %d, 响应体: %s", resp.StatusCode, string(body))
		sendErrorResponse(w, fmt.Sprintf("server returned non-200 status code: %d, body: %s", resp.StatusCode, string(body)), "internal_error", resp.StatusCode)
		return
	}

	log.Printf("开始处理响应流...")
	scanner := bufio.NewScanner(resp.Body)
	var allImageURLs []string

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("收到响应行: %s", line)

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event struct {
			Status  string `json:"status"`
			Payload []struct {
				Variations []struct {
					URL  string `json:"url"`
					IID  string `json:"iid"`
					Seed int    `json:"seed"`
				} `json:"variations"`
			} `json:"payload"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			log.Printf("解析事件失败: %v, data: %s", err, data)
			continue
		}

		// 收集所有图片URL
		for _, payload := range event.Payload {
			for _, variation := range payload.Variations {
				if variation.URL != "" {
					allImageURLs = append(allImageURLs, variation.URL)
					log.Printf("找到图片URL: %s", variation.URL)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("错误: 读取响应流失败: %v", err)
		sendErrorResponse(w, fmt.Sprintf("error reading response stream: %v", err), "internal_error", http.StatusInternalServerError)
		return
	}

	if len(allImageURLs) == 0 {
		log.Printf("错误: 未找到有效的图片URL")
		sendErrorResponse(w, "no valid image URLs found", "internal_error", http.StatusInternalServerError)
		return
	}

	log.Printf("成功获取图片URL: %v", allImageURLs)

	// 构建OpenAI流式响应
	streamResp := OpenAIStreamResponse{
		ID:      generateUUID(),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   "dall-e-3",
		Choices: []Choice{
			{
				Delta: Delta{
					Content: fmt.Sprintf("生成的图片:\n1. ![image](%s)\n2. ![image](%s)", allImageURLs[0], allImageURLs[1]),
					Role:    "assistant",
				},
				Index: 0,
			},
		},
	}

	respBytes, err := json.Marshal(streamResp)
	if err != nil {
		log.Printf("错误: 序列化响应失败: %v", err)
		sendErrorResponse(w, fmt.Sprintf("marshal response failed: %v", err), "internal_error", http.StatusInternalServerError)
		return
	}

	// 发送数据
	fmt.Fprintf(w, "data: %s\n\n", string(respBytes))
	flusher.Flush()

	// 发送结束标记
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	log.Printf("响应发送完成")
}

func processCompletedEvent(event struct {
	Status string `json:"status"`
	Data   struct {
		Content   string `json:"content"`
		EventType string `json:"eventType"`
		Error     string `json:"error,omitempty"`
	} `json:"data"`
}) ([]struct {
	URL     string `json:"url"`
	B64JSON string `json:"b64_json,omitempty"`
}, error) {
	var imageURLs []struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64_json,omitempty"`
	}

	if event.Data.Content == "" {
		return nil, fmt.Errorf("空content响应")
	}

	imageURLs = append(imageURLs, struct {
		URL     string `json:"url"`
		B64JSON string `json:"b64_json,omitempty"`
	}{
		URL:     event.Data.Content,
		B64JSON: "",
	})

	return imageURLs, nil
}

func handleImageResponse(w http.ResponseWriter, urls []string) error {
	if len(urls) == 0 {
		log.Printf("No URLs provided to handleImageResponse")
		return fmt.Errorf("no image URLs available")
	}

	log.Printf("Handling image response with URL: %s", urls[0])

	// 构造标准的 OpenAI 图片响应格式
	response := struct {
		Created int64 `json:"created"`
		Data    []struct {
			URL string `json:"url"`
		} `json:"data"`
	}{
		Created: time.Now().Unix(),
		Data: []struct {
			URL string `json:"url"`
		}{{URL: urls[0]}},
	}

	// 设置标准的 OpenAI 响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("x-request-id", generateUUID())
	w.Header().Set("openai-model", "dall-e-3")
	w.Header().Set("openai-organization", "org-default")
	w.Header().Set("openai-processing-ms", "3547")
	w.Header().Set("openai-version", "2020-10-01")
	w.Header().Set("x-ratelimit-limit-requests", "50")
	w.Header().Set("x-ratelimit-remaining-requests", "49")
	w.Header().Set("x-ratelimit-reset-requests", "1714520399")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		return fmt.Errorf("failed to encode response: %v", err)
	}

	log.Printf("Successfully sent image response with URL: %s", urls[0])
	return nil
}

func writeStreamResponse(w http.ResponseWriter, flusher http.Flusher, content string, isLast bool) error {
	// 设置必要的响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("x-request-id", generateUUID())

	response := OpenAIStreamResponse{
		ID:      "chatcmpl-" + generateUUID(),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []Choice{
			{
				Delta: struct {
					Content string `json:"content,omitempty"`
					Role    string `json:"role,omitempty"`
				}{
					Content: content,
					Role:    "assistant",
				},
				Index:        0,
				FinishReason: "",
			},
		},
	}

	if isLast {
		response.Choices[0].FinishReason = "stop"
	}

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}

	if isLast {
		if _, err := fmt.Fprintf(w, "data: [DONE]\n\n"); err != nil {
			return err
		}
	}

	flusher.Flush()
	return nil
}

func streamFromMerlin(merlinReq MerlinRequest, w http.ResponseWriter, flusher http.Flusher) (string, error) {
	// 设置响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("x-request-id", generateUUID())

	// 获取 session token
	sessionToken := os.Getenv("MERLIN_SESSION_TOKEN")
	if sessionToken == "" {
		return "", fmt.Errorf("MERLIN_SESSION_TOKEN is not set")
	}

	// 构建 session 请求
	sessionReq, err := http.NewRequest("GET", "https://session.getmerlin.in/?from=web", nil)
	if err != nil {
		return "", fmt.Errorf("create session request failed: %v", err)
	}

	// 设置请求头
	sessionReq.Header.Set("accept", "application/json, text/plain, */*")
	sessionReq.Header.Set("accept-language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	sessionReq.Header.Set("cache-control", "no-cache")
	sessionReq.Header.Set("origin", "https://www.getmerlin.in")
	sessionReq.Header.Set("pragma", "no-cache")
	sessionReq.Header.Set("referer", "https://www.getmerlin.in/")
	sessionReq.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	sessionReq.Header.Set("sec-ch-ua-mobile", "?0")
	sessionReq.Header.Set("sec-ch-ua-platform", `"macOS"`)
	sessionReq.Header.Set("sec-fetch-dest", "empty")
	sessionReq.Header.Set("sec-fetch-mode", "cors")
	sessionReq.Header.Set("sec-fetch-site", "same-site")
	sessionReq.Header.Set("cookie", fmt.Sprintf("__Secure-authjs.session-token=%s", sessionToken))

	log.Printf("Sending request to get session token...")
	client := &http.Client{}
	sessionResp, err := client.Do(sessionReq)
	if err != nil {
		return "", fmt.Errorf("session request failed: %v", err)
	}
	defer sessionResp.Body.Close()

	body, err := io.ReadAll(sessionResp.Body)
	if err != nil {
		return "", fmt.Errorf("read session response failed: %v", err)
	}
	log.Printf("Session response: %s", string(body))

	var sessionData struct {
		User struct {
			AccessToken string `json:"accessToken"`
		} `json:"user"`
	}
	if err := json.Unmarshal(body, &sessionData); err != nil {
		return "", fmt.Errorf("decode session response failed: %v, response: %s", err, string(body))
	}

	if sessionData.User.AccessToken == "" {
		return "", fmt.Errorf("received empty access token, response: %s", string(body))
	}

	token := sessionData.User.AccessToken
	log.Printf("Successfully got access token")

	// 发送聊天请求
	merlinReqBody, err := json.Marshal(merlinReq)
	if err != nil {
		return "", fmt.Errorf("marshal request body failed: %v", err)
	}

	log.Printf("Sending request to Merlin: %s", string(merlinReqBody))

	chatReq, err := http.NewRequest("POST", "https://arcane.getmerlin.in/v1/thread/unified", strings.NewReader(string(merlinReqBody)))
	if err != nil {
		return "", fmt.Errorf("create chat request failed: %v", err)
	}

	// 设置聊天请求头
	chatReq.Header.Set("Content-Type", "application/json")
	chatReq.Header.Set("Accept", "text/event-stream")
	chatReq.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	chatReq.Header.Set("Origin", "https://www.getmerlin.in")
	chatReq.Header.Set("Referer", "https://www.getmerlin.in/")
	chatReq.Header.Set("x-merlin-version", "web-merlin")
	chatReq.Header.Set("priority", "u=1, i")
	chatReq.Header.Set("cache-control", "no-cache")
	chatReq.Header.Set("pragma", "no-cache")
	chatReq.Header.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	chatReq.Header.Set("sec-ch-ua-mobile", "?0")
	chatReq.Header.Set("sec-ch-ua-platform", `"macOS"`)
	chatReq.Header.Set("Sec-Fetch-Dest", "empty")
	chatReq.Header.Set("Sec-Fetch-Mode", "cors")
	chatReq.Header.Set("Sec-Fetch-Site", "same-site")
	chatReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	chatReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	log.Printf("Final Request Headers:")
	for name, values := range chatReq.Header {
		if len(values) > 0 {
			log.Printf("%s: %s", name, values[0])
		}
	}

	resp, err := client.Do(chatReq)
	if err != nil {
		return "", fmt.Errorf("chat request failed: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Merlin response status: %s", resp.Status)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Received line: %s", line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		var event struct {
			Status string `json:"status"`
			Data   struct {
				Content     string `json:"content"`
				EventType   string `json:"eventType"`
				Attachments []struct {
					Type string `json:"type"`
					URL  string `json:"url"`
				} `json:"attachments"`
			} `json:"data"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			log.Printf("Error parsing event: %v, data: %s", err, data)
			continue
		}

		// 只处理实际的内容消息
		if event.Data.Content != "" && event.Status != "system" {
			response := OpenAIStreamResponse{
				ID:      "chatcmpl-" + generateUUID(),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   merlinReq.Model,
				Choices: []Choice{
					{
						Delta: struct {
							Content string `json:"content,omitempty"`
							Role    string `json:"role,omitempty"`
						}{
							Content: event.Data.Content,
						},
						Index: 0,
					},
				},
			}

			data, err := json.Marshal(response)
			if err != nil {
				continue
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return "", err
			}
			flusher.Flush()
		}

		// 检查是否为完成事件
		if event.Status == "system" && event.Data.EventType == "DONE" {
			// 发送最后的 [DONE] 消息
			if _, err := fmt.Fprintf(w, "data: [DONE]\n\n"); err != nil {
				return "", err
			}
			flusher.Flush()
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read response failed: %v", err)
	}

	return "", nil
}

func sendToMerlin(merlinReq MerlinRequest) (string, error) {
	var err error

	// 获取 token
	token, err := getTokenWithCache()
	if err != nil {
		return "", fmt.Errorf("error getting token: %v", err)
	}

	// 发送聊天请求
	merlinReq.Language = "CHINESE_SIMPLIFIED"
	merlinReq.Mode = "UNIFIED_CHAT"
	merlinReq.Metadata.WebAccess = true

	merlinReqBody, err := json.Marshal(merlinReq)
	if err != nil {
		return "", fmt.Errorf("marshal request body failed: %v", err)
	}

	log.Printf("Sending request to Merlin: %s", string(merlinReqBody))

	chatReq, err := http.NewRequest("POST", "https://arcane.getmerlin.in/v1/thread/unified", strings.NewReader(string(merlinReqBody)))
	if err != nil {
		return "", fmt.Errorf("create chat request failed: %v", err)
	}

	// 设置聊天请求头
	chatReq.Header.Set("Content-Type", "application/json")
	chatReq.Header.Set("Accept", "text/event-stream")
	chatReq.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	chatReq.Header.Set("Origin", "https://www.getmerlin.in")
	chatReq.Header.Set("Referer", "https://www.getmerlin.in/")
	chatReq.Header.Set("x-merlin-version", "web-merlin")
	chatReq.Header.Set("priority", "u=1, i")
	chatReq.Header.Set("cache-control", "no-cache")
	chatReq.Header.Set("pragma", "no-cache")
	chatReq.Header.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	chatReq.Header.Set("sec-ch-ua-mobile", "?0")
	chatReq.Header.Set("sec-ch-ua-platform", `"macOS"`)
	chatReq.Header.Set("Sec-Fetch-Dest", "empty")
	chatReq.Header.Set("Sec-Fetch-Mode", "cors")
	chatReq.Header.Set("Sec-Fetch-Site", "same-site")
	chatReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	chatReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	log.Printf("Final Request Headers:")
	for name, values := range chatReq.Header {
		if len(values) > 0 {
			log.Printf("%s: %s", name, values[0])
		}
	}

	client := &http.Client{}
	resp, err := client.Do(chatReq)
	if err != nil {
		return "", fmt.Errorf("chat request failed: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Merlin response status: %s", resp.Status)

	var imageUrls []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Received line: %s", line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		var event struct {
			Status string `json:"status"`
			Data   struct {
				Content   string `json:"content"`
				EventType string `json:"eventType"`
				Message   struct {
					Attachments []struct {
						Type string `json:"type"`
						URL  string `json:"url"`
					} `json:"attachments"`
				} `json:"message"`
			} `json:"data"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			log.Printf("Error parsing event: %v, data: %s", err, data)
			continue
		}

		// 检查 Data.Message.Attachments
		for _, attachment := range event.Data.Message.Attachments {
			if attachment.URL != "" && attachment.Type == "IMAGE" {
				log.Printf("Found URL in Data.Message.Attachments: %s", attachment.URL)
				imageUrls = append(imageUrls, attachment.URL)
			}
		}

		// 检查是否为完成事件
		if event.Status == "system" && event.Data.EventType == "DONE" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read response failed: %v", err)
	}

	if len(imageUrls) == 0 {
		log.Printf("No image URLs found in response")
		return "", fmt.Errorf("no valid image URLs generated")
	}

	log.Printf("Found %d image URLs: %v", len(imageUrls), imageUrls)
	return imageUrls[0], nil
}

func sendErrorResponse(w http.ResponseWriter, message string, errorType string, statusCode int) {
	// 构造标准的 OpenAI 错误响应
	errorResponse := struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Param   string `json:"param,omitempty"`
			Code    string `json:"code,omitempty"`
		} `json:"error"`
	}{
		Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Param   string `json:"param,omitempty"`
			Code    string `json:"code,omitempty"`
		}{
			Message: message,
			Type:    errorType,
			Code:    "error",
		},
	}

	// 设置标准的 OpenAI 错误响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("x-request-id", generateUUID())
	w.Header().Set("openai-version", "2020-10-01")
	w.Header().Set("openai-organization", "org-default")
	w.Header().Set("openai-processing-ms", "50")
	w.Header().Set("x-ratelimit-limit-requests", "50")
	w.Header().Set("x-ratelimit-remaining-requests", "49")
	w.Header().Set("x-ratelimit-reset-requests", "1714520399")

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

func HandleChat(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理 OPTIONS 请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("Received request to path: %s", r.URL.Path)

	if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"GetMerlin2Api Service Running...","message":"MoLoveSze..."}`)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		sendErrorResponse(w, "Method not allowed", "invalid_request_error", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to decode request: %v", err), "invalid_request_error", http.StatusBadRequest)
		return
	}

	// 设置默认模型为 gpt-4o-64k-output
	if req.Model == "" {
		req.Model = "gpt-4o-64k-output"
	}

	// 清理历史消息，只保留最后一条
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		req.Messages = []Message{lastMsg}
	}

	log.Printf("Processing request: %+v", req)

	if len(req.Messages) == 0 {
		sendErrorResponse(w, "No messages in request", "invalid_request_error", http.StatusBadRequest)
		return
	}

	lastMsg := req.Messages[0]
	log.Printf("Processing message with model: %s", req.Model)
	log.Printf("Message content: %s", lastMsg.Content)

	// 检查是否为图片生成请求
	if req.Model == "flux-1.1-pro" || req.Model == "recraft-v3" {
		// 将图片生成请求重定向到标准的 OpenAI 图片生成接口
		imageReq := OpenAIImageGenerationRequest{
			Prompt: lastMsg.Content,
			N:      1,
			Size:   "1024x1024",
			Model:  req.Model,
		}

		// 获取 flusher
		flusher, ok := w.(http.Flusher)
		if !ok {
			sendErrorResponse(w, "Streaming unsupported!", "internal_error", http.StatusInternalServerError)
			return
		}

		generateImage(w, flusher, imageReq.Prompt, req.Model)
		return
	}

	chatID := uuid.New().String()
	messageID := uuid.New().String()
	childID := uuid.New().String()

	merlinReq := MerlinRequest{
		Attachments: []interface{}{},
		ChatID:      chatID,
		Language:    "CHINESE_SIMPLIFIED",
		Message: struct {
			Content  string `json:"content"`
			Context  string `json:"context"`
			ChildID  string `json:"childId"`
			ID       string `json:"id"`
			ParentID string `json:"parentId"`
		}{
			Content:  lastMsg.Content,
			Context:  "",
			ChildID:  childID,
			ID:       messageID,
			ParentID: "root",
		},
		Metadata: struct {
			LargeContext  bool `json:"largeContext"`
			MerlinMagic   bool `json:"merlinMagic"`
			ProFinderMode bool `json:"proFinderMode"`
			WebAccess     bool `json:"webAccess"`
			Temperature   int  `json:"temperature"`
		}{
			LargeContext:  false,
			MerlinMagic:   false,
			ProFinderMode: false,
			WebAccess:     true,
			Temperature:   0,
		},
		Mode:  "UNIFIED_CHAT",
		Model: req.Model,
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// 发送初始消息
		response := OpenAIStreamResponse{
			ID:      "chatcmpl-" + generateUUID(),
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []Choice{
				{
					Delta: struct {
						Content string `json:"content,omitempty"`
						Role    string `json:"role,omitempty"`
					}{
						Role: "assistant",
					},
					Index: 0,
				},
			},
		}

		data, err := json.Marshal(response)
		if err == nil {
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		content, err := streamFromMerlin(merlinReq, w, flusher)
		if err != nil {
			log.Printf("Error streaming from Merlin: %v", err)
			return
		}
		log.Printf("Streaming completed: %s", content)
	} else {
		content, err := sendToMerlin(merlinReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to send request to Merlin: %v", err), http.StatusInternalServerError)
			return
		}

		response := ChatResponse{
			ID:      messageID,
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []Choice{
				{
					Index: 0,
					Delta: Delta{
						Content: content,
					},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func HandleImageGeneration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 解析请求体
	var req MerlinImageGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode request: %v", err), http.StatusBadRequest)
		return
	}

	// 创建响应写入器
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// 生成图片
	generateImage(w, flusher, req.Action.Message.Content, req.Action.Message.Metadata.Context)
}

func HandleImageGenerations(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理 OPTIONS 请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", "invalid_request_error", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求
	var req OpenAIImageGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body", "invalid_request_error", http.StatusBadRequest)
		return
	}

	// 验证必需字段
	if req.Prompt == "" {
		sendErrorResponse(w, "Prompt is required", "invalid_request_error", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if req.N == 0 {
		req.N = 1
	}
	if req.Size == "" {
		req.Size = "1024x1024"
	}
	if req.Model == "" {
		req.Model = "dall-e-3"
	}

	// 获取 flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		sendErrorResponse(w, "Streaming unsupported!", "internal_error", http.StatusInternalServerError)
		return
	}

	// 生成图片
	generateImage(w, flusher, req.Prompt, req.Model)
}
