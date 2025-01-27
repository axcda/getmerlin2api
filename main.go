package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rubleowen/GetMerlin2Api/api"
	"github.com/rubleowen/GetMerlin2Api/auth"
	"github.com/rubleowen/GetMerlin2Api/utils"
)

func main() {
	utils.LoadEnv()
	// 生成一个token用于测试
	token, err := auth.GenerateToken()
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}
	// 打印token的前100个字符，避免日志过长
	fmt.Printf("Token prefix: %s...\n", token[:100])

	// 注册路由
	http.HandleFunc("/", api.HandleChat)
	http.HandleFunc("/v1/chat/completions", api.HandleChat)
	http.HandleFunc("/v1/images/generations", api.HandleImageGeneration)

	// 启动服务器
	port := "8081"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
