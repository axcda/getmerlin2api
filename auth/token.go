package auth

import (
	"fmt"
	"strings"

	"github.com/rubleowen/GetMerlin2Api/utils"
)

// GenerateToken 获取认证token
func GenerateToken() (string, error) {
	token := utils.GetEnvOrDefault("MERLIN_TOKEN", "")
	if token == "" {
		return "", fmt.Errorf("MERLIN_TOKEN 环境变量未设置")
	}

	// 移除可能存在的多余空格
	token = strings.TrimSpace(token)

	// 如果token已经包含Bearer前缀，直接返回
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return token, nil
	}

	// 如果没有Bearer前缀，添加它
	return "Bearer " + token, nil
}
