# GetMerlin2Api

这是一个将 Merlin API 转换为 OpenAI API 格式的代理服务器，支持聊天和图片生成功能。

## 功能特点

- ✨ 支持聊天对话（Chat Completions）
- 🎨 支持图片生成（Image Generation）
- 🔄 支持流式响应（Stream Response）
- 🔌 完全兼容 OpenAI API 格式
- 🔑 自动处理 Merlin 认证
- 🔁 支持 Session Token 认证
- 🚀 高性能，低延迟
- 🔒 安全的令牌管理
- 💡 简单易用的配置

## 获取 Session Token

1. 登录 Merlin 网站后，打开浏览器开发者工具（F12 或右键检查）
2. 在开发者工具中切换到"网络/Network"标签
3. 找到对 `https://session.getmerlin.in/?from=web` 的 GET 请求
4. 在请求响应中可以找到 Session Token

获取步骤图解：
1. 打开 Merlin 网站并登录
2. 按 F12 打开开发者工具
3. 点击 Network 标签
4. 找到 session.getmerlin.in 请求
5. 查看响应内容获取 token

> 提示：Session Token 通常会定期过期，建议在过期前及时更新。

## 技术要求

- Go 1.16 或更高版本
- 有效的 Merlin Session Token
- 系统内存：至少 2GB RAM
- 磁盘空间：至少 100MB 可用空间

## 依赖项

- github.com/google/uuid v1.6.0
- github.com/joho/godotenv v1.5.1
- github.com/json-iterator/go v1.1.12
- github.com/patrickmn/go-cache v2.1.0+incompatible

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
创建 `.env` 文件并添加你的认证信息：

```bash
MERLIN_SESSION_TOKEN=your_session_token_here
PORT=8081  # 可选，默认为 8081
```

> 注意：请确保 Session Token 的有效性，如果 Token 过期需要手动更新。

4. 运行服务：
```bash
go run main.go
```

服务默认运行在 `8081` 端口。

## API 使用说明

### 支持的模型

#### 聊天模型
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

#### 画图模型
- flux-1.1-pro
- recraft-v3
### API 端点

#### 1. 聊天对话

- 端点：`/v1/chat/completions`
- 方法：`POST`
- 支持流式输出：`是`

示例请求：
```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "你好"}],
    "stream": true
  }'
```

#### 2. 图片生成

- 端点：`/v1/chat/completions`
- 方法：`POST`
- 支持流式输出：`是`

示例请求：
```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "flux-1.1-pro",
    "messages": [{"role": "user", "content": "一只可爱的猫"}],
    "stream": true
  }'
```

## 在第三方应用中使用

### OpenWebUI/Cherry Studio 配置

1. 添加自定义模型：
   - 模型名称：`flux-1.1-pro`
   - API 地址：`http://localhost:8081`
   - API 密钥：（可选）

2. 选择 `flux-1.1-pro` 模型进行图片生成。

### 其他兼容 OpenAI API 的应用

本服务完全兼容 OpenAI API 格式，因此可以在任何支持 OpenAI API 的应用中使用，只需将 API 地址修改为：
```
http://localhost:8081
```


## 常见问题

1. Token 过期问题
   - Session Token 过期时需要手动更新
   - 按照上述"获取 Session Token"步骤重新获取
   - 通常在 `https://session.getmerlin.in/?from=web` 接口可获取最新 token

2. 并发限制
   - 默认支持多并发请求
   - 可通过环境变量调整限制

3. 网络问题
   - 确保网络稳定
   - 检查防火墙设置

## 安全建议

- 不要在公网环境直接暴露服务
- 定期更新依赖包
- 使用 HTTPS 进行传输
- 妥善保管 Session Token 信息
- 定期更新 Session Token

## 贡献指南

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/AmazingFeature`
3. 提交更改：`git commit -m 'Add some AmazingFeature'`
4. 推送分支：`git push origin feature/AmazingFeature`
5. 提交 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 联系方式

如有问题或建议，欢迎提交 Issue 或 Pull Request！