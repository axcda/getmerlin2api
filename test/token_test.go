package test

import (
	"testing"

	"github.com/rubleowen/GetMerlin2Api/auth"
	"github.com/rubleowen/GetMerlin2Api/utils"
)

func TestRefreshAuthToken(t *testing.T) {
	// 从环境变量获取refresh token
	refreshToken := utils.GetEnvOrDefault("MERLIN_REFRESH_TOKEN", "")
	if refreshToken == "" {
		t.Fatal("MERLIN_REFRESH_TOKEN environment variable is not set")
	}

	// 调用RefreshAuthToken函数
	token, err := auth.RefreshAuthToken(refreshToken)
	if err != nil {
		t.Fatalf("RefreshAuthToken failed: %v", err)
	}

	// 验证返回的token
	if token == "" {
		t.Error("Received empty token")
	}

	if len(token) < 20 {
		t.Error("Token seems too short to be valid")
	}

	// 验证token格式
	if token[:7] != "Bearer " {
		t.Error("Token should start with 'Bearer '")
	}
}
