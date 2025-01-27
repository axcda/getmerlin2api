# GetMerlin2Api

这是一个将 Merlin API 转换为 OpenAI API 格式的代理服务器，支持聊天和图片生成功能。

## 功能特点

- ✨ 支持聊天对话（Chat Completions）
- 🎨 支持图片生成（Image Generation）
- 🔄 支持流式响应（Stream Response）
- 🔌 完全兼容 OpenAI API 格式
- 🔑 自动处理 Merlin 认证

## 安装要求

- Go 1.16 或更高版本
- 有效的 Merlin API Token

## 快速开始

1. 克隆仓库：
```bash
git clone https://github.com/axcda/getmerlin2api.git
cd GetMerlin2Api
```

2. 安装依赖：
```bash
go mod download
```

3. 配置环境变量：
创建 `.env` 文件并添加你的 Merlin Token：
```bash
MERLIN_TOKEN=your_merlin_token_here
```

4. 运行服务：
```bash
go run main.go
```

服务默认运行在 `8081` 端口。

## API 使用说明
### 支持的模型
```
o1 mini
DeepSeek R1
Claude 3.5 Sonnet
DeepSeek v3
Gemini 1.5 Pro
GPT 4o
Llama 3.1 405B
Claude 3.5 Haiku
Claude 3 Haiku
Gemini 1.5
GPT 4o Mini
GPT 4o （Longer Output）
o1
o1 Preview
```
### 画图模型
我这里就测试了个flux-1.1-pro
### 聊天对话

```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "你好"}],
    "stream": true
  }'
```

### 图片生成

```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "flux-1.1-pro",
    "messages": [{"role": "user", "content": "一只可爱的猫"}],
    "stream": true
  }'
```

## 在 OpenWebUI/Cherry Studio 中使用

1. 添加自定义模型：
   - 模型名称：`flux-1.1-pro`
   - API 地址：`http://localhost:8081`
   - API 密钥：（可选）

2. 选择 `flux-1.1-pro` 模型进行图片生成。

## 注意事项

- 确保 `MERLIN_TOKEN` 环境变量已正确设置
- 图片生成请求需要使用 `flux-1.1-pro` 作为模型名称
- 响应格式完全兼容 OpenAI API，可以直接在支持 OpenAI 的客户端中使用


## 贡献

欢迎提交 Issue 和 Pull Request！