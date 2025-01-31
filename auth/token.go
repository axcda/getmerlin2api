package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/rubleowen/GetMerlin2Api/utils"
)

var (
	cachedToken     string
	cachedExpiry    time.Time
	tokenLock       sync.Mutex
	refreshInterval = 55 * time.Minute // Token通常1小时过期，提前5分钟刷新
)

type SessionResponse struct {
	User struct {
		AccessToken string `json:"accessToken"`
		Email       string `json:"email"`
		Name        string `json:"name"`
	} `json:"user"`
	Expires string `json:"expires"`
}

type RefreshResponse struct {
	Status string `json:"status"`
	Data   struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	} `json:"data"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// GetSessionToken 从session.getmerlin.in获取token
func GetSessionToken(sessionToken string) (string, error) {
	log.Printf("Trying to get session token...")
	req, err := http.NewRequest("GET", "https://session.getmerlin.in/?from=web", nil)
	if err != nil {
		return "", fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("origin", "https://www.getmerlin.in")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("referer", "https://www.getmerlin.in/")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("x-merlin-version", "web-merlin")
	req.Header.Set("cookie", fmt.Sprintf("__Secure-authjs.session-token=%s", sessionToken))

	log.Printf("Sending request to session.getmerlin.in...")
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %v", err)
	}

	log.Printf("Response from session.getmerlin.in: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get session token failed: %s", string(body))
	}

	var sessionResp SessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %v", err)
	}

	if sessionResp.User.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	// 更新缓存
	token := sessionResp.User.AccessToken
	cachedToken = token
	cachedExpiry = time.Now().Add(refreshInterval)

	log.Printf("Successfully got session token")
	return token, nil
}

// RefreshAuthToken 通过refresh token获取新的authorization token
func RefreshAuthToken(refreshToken string) (string, error) {
	log.Printf("Trying to refresh token...")
	tokenLock.Lock()
	defer tokenLock.Unlock()

	// 如果缓存的token还有效，直接返回
	if cachedToken != "" && time.Now().Before(cachedExpiry) {
		log.Printf("Using cached token")
		return cachedToken, nil
	}

	// 准备请求体
	reqBody := map[string]interface{}{
		"token": refreshToken,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request body failed: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://uam.getmerlin.in/session/get", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.getmerlin.in")
	req.Header.Set("Referer", "https://www.getmerlin.in/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("x-merlin-version", "web-merlin")
	req.Header.Set("x-merlin-client-type", "web")
	req.Header.Set("x-merlin-client-version", "1.0.0")
	req.Header.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")

	// 设置Authorization header
	req.Header.Set("Authorization", refreshToken)

	log.Printf("Sending request to uam.getmerlin.in...")
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %v", err)
	}

	log.Printf("Response from uam.getmerlin.in: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("refresh token failed: %s", string(body))
	}

	var refreshResp RefreshResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %v", err)
	}

	if refreshResp.Status == "error" {
		return "", fmt.Errorf("refresh token failed: %s", string(body))
	}

	if refreshResp.Data.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	// 更新缓存
	token := refreshResp.Data.AccessToken
	cachedToken = token
	cachedExpiry = time.Now().Add(refreshInterval)

	log.Printf("Successfully refreshed token")
	return token, nil
}

// GenerateToken 获取认证token
func GenerateToken() (string, error) {
	log.Printf("Generating token...")
	// 优先使用 session token
	sessionToken := utils.GetEnvOrDefault("MERLIN_SESSION_TOKEN", "")
	if sessionToken != "" {
		log.Printf("Using session token")
		token, err := GetSessionToken(sessionToken)
		if err == nil {
			return token, nil
		}
		log.Printf("Session token failed: %v", err)
	}

	// 尝试使用refresh token
	refreshToken := utils.GetEnvOrDefault("MERLIN_REFRESH_TOKEN", "")
	if refreshToken != "" {
		log.Printf("Using refresh token")
		token, err := RefreshAuthToken(refreshToken)
		if err == nil {
			return token, nil
		}
		log.Printf("Refresh token failed: %v", err)
	}

	// 最后才尝试使用普通token
	token := utils.GetEnvOrDefault("MERLIN_TOKEN", "")
	if token == "" {
		return "", fmt.Errorf("no valid token found in environment variables")
	}

	log.Printf("Using normal token")
	return token, nil
}
