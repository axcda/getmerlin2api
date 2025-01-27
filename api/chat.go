package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubleowen/GetMerlin2Api/auth"
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

type MerlinRequest struct {
	Attachments []interface{} `json:"attachments"`
	ChatId      string        `json:"chatId"`
	Language    string        `json:"language"`
	Message     struct {
		Content  string `json:"content"`
		Context  string `json:"context"`
		ChildId  string `json:"childId"`
		Id       string `json:"id"`
		ParentId string `json:"parentId"`
	} `json:"message"`
	Metadata struct {
		LargeContext  bool `json:"largeContext"`
		MerlinMagic   bool `json:"merlinMagic"`
		ProFinderMode bool `json:"proFinderMode"`
		WebAccess     bool `json:"webAccess"`
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
			Content string `json:"content"`
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
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
		URL string `json:"url"`
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

func getToken() (string, error) {
	return auth.GenerateToken()
}

func HandleChat(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request to path: %s", r.URL.Path)

	if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"GetMerlin2Api Service Running...","message":"MoLoveSze..."}`)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var openAIReq OpenAIRequest
	if err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&openAIReq); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received request: %+v\n", openAIReq)

	if len(openAIReq.Messages) == 0 {
		log.Printf("Empty messages array in request")
		http.Error(w, "Messages array cannot be empty", http.StatusBadRequest)
		return
	}

	lastMessage := openAIReq.Messages[len(openAIReq.Messages)-1]
	log.Printf("Processing message with model: %s", openAIReq.Model)
	log.Printf("Message content: %s", lastMessage.Content)

	if strings.Contains(strings.ToLower(openAIReq.Model), "flux-1.1-pro") {
		log.Printf("Detected image generation model, forwarding to image generation handler")

		// 创建一个管道来接收图片生成的响应
		pr, pw := io.Pipe()

		// 在新的goroutine中处理图片生成
		go func() {
			defer pw.Close()
			imageReq := &http.Request{
				Method: http.MethodPost,
				Body:   io.NopCloser(strings.NewReader(fmt.Sprintf(`{"prompt":"%s","n":1,"size":"1024x1024"}`, lastMessage.Content))),
				Header: r.Header,
			}

			responseWriter := &responseWriter{pw}
			HandleImageGeneration(responseWriter, imageReq)
		}()

		// 读取图片生成的响应
		var imageResp ImageGenerationResponse
		if err := json.NewDecoder(pr).Decode(&imageResp); err != nil {
			log.Printf("Error decoding image response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if openAIReq.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			// 发送处理中的消息
			processingResp := OpenAIResponse{
				Id:      generateUUID(),
				Object:  "chat.completion.chunk",
				Created: getCurrentTimestamp(),
				Model:   openAIReq.Model,
				Choices: []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					Index        int    `json:"index"`
					FinishReason string `json:"finish_reason"`
				}{{
					Delta: struct {
						Content string `json:"content"`
					}{
						Content: "正在生成图片...",
					},
					Index:        0,
					FinishReason: "",
				}},
			}
			respData, _ := json.Marshal(processingResp)
			fmt.Fprintf(w, "data: %s\n\n", string(respData))
			w.(http.Flusher).Flush()

			// 发送图片URL
			imageResp := OpenAIResponse{
				Id:      generateUUID(),
				Object:  "chat.completion.chunk",
				Created: getCurrentTimestamp(),
				Model:   openAIReq.Model,
				Choices: []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					Index        int    `json:"index"`
					FinishReason string `json:"finish_reason"`
				}{{
					Delta: struct {
						Content string `json:"content"`
					}{
						Content: fmt.Sprintf("![Generated Image](%s)", imageResp.Data[0].URL),
					},
					Index:        0,
					FinishReason: "stop",
				}},
			}
			respData, _ = json.Marshal(imageResp)
			fmt.Fprintf(w, "data: %s\n\n", string(respData))
			fmt.Fprintf(w, "data: [DONE]\n\n")
			w.(http.Flusher).Flush()
		} else {
			// 非流式响应
			response := map[string]interface{}{
				"id":      generateUUID(),
				"object":  "chat.completion",
				"created": getCurrentTimestamp(),
				"model":   openAIReq.Model,
				"choices": []map[string]interface{}{
					{
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": fmt.Sprintf("![Generated Image](%s)", imageResp.Data[0].URL),
						},
						"finish_reason": "stop",
						"index":         0,
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
		return
	}

	merlinReq := MerlinRequest{
		Attachments: make([]interface{}, 0),
		ChatId:      generateV1UUID(),
		Language:    "AUTO",
		Message: struct {
			Content  string `json:"content"`
			Context  string `json:"context"`
			ChildId  string `json:"childId"`
			Id       string `json:"id"`
			ParentId string `json:"parentId"`
		}{
			Content:  lastMessage.Content,
			Context:  "",
			ChildId:  generateUUID(),
			Id:       generateUUID(),
			ParentId: "root",
		},
		Mode:  "UNIFIED_CHAT",
		Model: openAIReq.Model,
		Metadata: struct {
			LargeContext  bool `json:"largeContext"`
			MerlinMagic   bool `json:"merlinMagic"`
			ProFinderMode bool `json:"proFinderMode"`
			WebAccess     bool `json:"webAccess"`
		}{
			LargeContext:  false,
			MerlinMagic:   false,
			ProFinderMode: false,
			WebAccess:     false,
		},
	}

	token, err := getToken()
	if err != nil {
		http.Error(w, "Failed to get token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Using token: %s\n", token)

	client := &http.Client{}
	merlinReqBody, _ := json.Marshal(merlinReq)

	log.Printf("Sending request to Merlin: %s\n", string(merlinReqBody))

	req, _ := http.NewRequest("POST", "https://arcane.getmerlin.in/v1/thread/unified", strings.NewReader(string(merlinReqBody)))

	log.Printf("Request Headers:")
	for key, values := range req.Header {
		log.Printf("%s: %v", key, values)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", token)
	req.Header.Set("Origin", "https://getmerlin.in")
	req.Header.Set("Referer", "https://getmerlin.in/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("x-merlin-version", "web-merlin")
	req.Header.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")

	log.Printf("Final Request Headers:")
	for key, values := range req.Header {
		log.Printf("%s: %v", key, values)
	}

	if openAIReq.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.Header().Set("Transfer-Encoding", "chunked")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to Merlin: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	log.Printf("Merlin response status: %s\n", resp.Status)

	if !openAIReq.Stream {
		var fullContent string
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Printf("Error reading line: %v\n", err)
				continue
			}

			line = strings.TrimSpace(line)
			log.Printf("Received line: %s\n", line)

			if strings.HasPrefix(line, "event: message") {
				dataLine, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("Error reading data line: %v\n", err)
					continue
				}
				dataLine = strings.TrimSpace(dataLine)
				log.Printf("Received data line: %s\n", dataLine)

				if strings.HasPrefix(dataLine, "data: ") {
					dataStr := strings.TrimPrefix(dataLine, "data: ")
					var merlinResp MerlinResponse
					if err := json.Unmarshal([]byte(dataStr), &merlinResp); err != nil {
						log.Printf("Error unmarshaling response: %v\n", err)
						continue
					}
					if merlinResp.Data.Content != " " && merlinResp.Data.Content != "" {
						fullContent += merlinResp.Data.Content
						log.Printf("Accumulated content: %s\n", fullContent)
					}
				}
			}
		}

		response := map[string]interface{}{
			"id":      generateUUID(),
			"object":  "chat.completion",
			"created": getCurrentTimestamp(),
			"model":   openAIReq.Model,
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": fullContent,
					},
					"finish_reason": "stop",
					"index":         0,
				},
			},
		}

		log.Printf("Sending response: %+v\n", response)
		json.NewEncoder(w).Encode(response)
		return
	}

	reader := bufio.NewReaderSize(resp.Body, 256)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "event: message") {
			dataLine, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Error reading data line: %v\n", err)
				continue
			}
			dataLine = strings.TrimSpace(dataLine)

			if strings.HasPrefix(dataLine, "data: ") {
				dataStr := strings.TrimPrefix(dataLine, "data: ")
				var merlinResp MerlinResponse
				if err := json.Unmarshal([]byte(dataStr), &merlinResp); err != nil {
					log.Printf("Error unmarshaling response: %v\n", err)
					continue
				}

				if merlinResp.Data.Content != "" {
					openAIResp := OpenAIResponse{
						Id:      generateUUID(),
						Object:  "chat.completion.chunk",
						Created: getCurrentTimestamp(),
						Model:   openAIReq.Model,
						Choices: []struct {
							Delta struct {
								Content string `json:"content"`
							} `json:"delta"`
							Index        int    `json:"index"`
							FinishReason string `json:"finish_reason"`
						}{{
							Delta: struct {
								Content string `json:"content"`
							}{
								Content: merlinResp.Data.Content,
							},
							Index:        0,
							FinishReason: "",
						}},
					}

					respData, _ := json.Marshal(openAIResp)
					fmt.Fprintf(w, "data: %s\n\n", string(respData))
					w.(http.Flusher).Flush()
				}
			}
		}
	}

	finalResp := OpenAIResponse{
		Choices: []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
			Index        int    `json:"index"`
			FinishReason string `json:"finish_reason"`
		}{{
			Delta: struct {
				Content string `json:"content"`
			}{Content: ""},
			Index:        0,
			FinishReason: "stop",
		}},
	}
	respData, _ := json.Marshal(finalResp)
	fmt.Fprintf(w, "data: %s\n\n", string(respData))
	fmt.Fprintf(w, "data: [DONE]\n\n")
	w.(http.Flusher).Flush()
}

func HandleImageGeneration(w http.ResponseWriter, r *http.Request) {
	log.Printf("Starting image generation handler")

	var req ImageGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding image generation request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Image generation request: prompt=%s, n=%d, size=%s", req.Prompt, req.N, req.Size)

	merlinReq := MerlinImageRequest{
		Action: struct {
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
		}{
			Message: struct {
				Attachments []interface{} `json:"attachments"`
				Content     string        `json:"content"`
				Metadata    struct {
					Context string `json:"context"`
				} `json:"metadata"`
				ParentId string `json:"parentId"`
				Role     string `json:"role"`
			}{
				Attachments: []interface{}{},
				Content:     req.Prompt,
				Metadata: struct {
					Context string `json:"context"`
				}{
					Context: "",
				},
				ParentId: uuid.New().String(),
				Role:     "user",
			},
			Type: "NEW",
		},
		ChatId: uuid.New().String(),
		Mode:   "IMAGE_CHAT",
		Settings: struct {
			MerlinPromptMagic bool `json:"merlinPromptMagic"`
			ModelConfig       []struct {
				AspectRatio    string `json:"aspectRatio"`
				ModelId        string `json:"modelId"`
				NumberOfImages int    `json:"numberOfImages"`
			} `json:"modelConfig"`
			NegativePrompt string `json:"negativePrompt"`
		}{
			MerlinPromptMagic: false,
			ModelConfig: []struct {
				AspectRatio    string `json:"aspectRatio"`
				ModelId        string `json:"modelId"`
				NumberOfImages int    `json:"numberOfImages"`
			}{
				{
					AspectRatio:    "1:1",
					ModelId:        "black-forest-labs/flux-1.1-pro",
					NumberOfImages: req.N,
				},
			},
			NegativePrompt: "",
		},
	}

	log.Printf("Prepared Merlin request for image generation")

	client := &http.Client{}
	jsonData, err := json.Marshal(merlinReq)
	if err != nil {
		log.Printf("Error marshaling Merlin request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reqURL, _ := url.Parse("https://uam.getmerlin.in/web/v2/image-generation")
	httpReq, err := http.NewRequest("POST", reqURL.String(), strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("MERLIN_TOKEN")))

	log.Printf("Sending request to Merlin image generation API")
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("Error sending request to Merlin: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Error response from Merlin: %s", string(body))
		http.Error(w, fmt.Sprintf("Error from Merlin: %s", string(body)), resp.StatusCode)
		return
	}

	log.Printf("Successfully received response from Merlin")

	var imageUrls []string
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading from Merlin: %v", err)
			break
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "event: message") {
			dataLine, err := reader.ReadString('\n')
			if err != nil {
				continue
			}
			dataLine = strings.TrimSpace(dataLine)

			if strings.HasPrefix(dataLine, "data: ") {
				data := strings.TrimPrefix(dataLine, "data: ")
				var event struct {
					Status string `json:"status"`
					Data   struct {
						Attachments []struct {
							Type     string `json:"type"`
							URL      string `json:"url"`
							Metadata struct {
								Status string `json:"status"`
							} `json:"metadata"`
						} `json:"attachments"`
						Message struct {
							Attachments []struct {
								Type string `json:"type"`
								URL  string `json:"url"`
							} `json:"attachments"`
						} `json:"message"`
						EventType string `json:"eventType"`
						URL       string `json:"url"`
					} `json:"data"`
				}

				if err := json.Unmarshal([]byte(data), &event); err != nil {
					log.Printf("Error unmarshaling event: %v", err)
					log.Printf("Data: %s", data)
					continue
				}

				if event.Data.URL != "" {
					imageUrls = append(imageUrls, event.Data.URL)
					continue
				}

				for _, attachment := range event.Data.Attachments {
					if attachment.URL != "" && attachment.Type == "IMAGE" {
						imageUrls = append(imageUrls, attachment.URL)
					}
				}

				for _, attachment := range event.Data.Message.Attachments {
					if attachment.URL != "" && attachment.Type == "IMAGE" {
						imageUrls = append(imageUrls, attachment.URL)
					}
				}
			}
		}
	}

	if len(imageUrls) == 0 {
		http.Error(w, "No valid image URLs generated", http.StatusInternalServerError)
		return
	}

	response := ImageGenerationResponse{
		Created: time.Now().Unix(),
		Data: make([]struct {
			URL string `json:"url"`
		}, 1),
	}
	response.Data[0].URL = imageUrls[len(imageUrls)-1]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateUUID() string {
	return uuid.New().String()
}

func generateV1UUID() string {
	uuidObj := uuid.Must(uuid.NewUUID())
	return uuidObj.String()
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// 添加一个自定义的responseWriter来捕获响应
type responseWriter struct {
	io.Writer
}

func (w *responseWriter) Header() http.Header {
	return make(http.Header)
}

func (w *responseWriter) WriteHeader(statusCode int) {
}

func (w *responseWriter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}
