# Eino Minimal - 最小化AI聊天API

这是从原项目中抽离出来的最小化版本，专门用于集成eino框架的AI交流功能。

## 功能特性

- 基于eino框架的AI聊天功能
- 支持多种AI模型（千问、豆包、Ollama）
- 流式响应支持
- 对话历史管理
- 简单的HTTP API接口

## 项目结构

```
eino-minimal/
├── go.mod                    # 依赖管理
├── main.go                   # 主程序入口
├── README.md                 # 说明文档
├── models/
│   └── chat_model.go        # AI模型配置
├── memory/
│   └── simple_memory.go     # 对话内存管理
├── agent/
│   ├── agent.go             # 代理核心逻辑
│   ├── chat_template.go     # 聊天模板
│   └── tools.go             # 工具配置
└── api/
    └── chat_handler.go      # HTTP API处理器
```

## 快速开始

### 1. 安装依赖

```bash
cd eino-minimal
go mod tidy
```

### 2. 配置环境变量

根据你要使用的AI模型，设置相应的环境变量：

**千问模型（推荐）：**
```bash
export QWEN_MODEL="qwen-vl-plus"
export QWEN_API_KEY="your-qwen-api-key"
export QWEN_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
```

**豆包模型：**
```bash
export ARK_MODEL="your-ark-model"
export ARK_API_KEY="your-ark-api-key"
export ARK_BASE_URL="your-ark-base-url"
```

**Ollama本地模型：**
```bash
# 确保Ollama服务运行在localhost:11434
# 默认使用qwen3:latest模型
```

### 3. 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

### 4. 测试API

**聊天接口：**
```bash
curl "http://localhost:8080/api/chat?id=test-conversation&message=你好"
```

**健康检查：**
```bash
curl "http://localhost:8080/health"
```

## API 接口

### GET /api/chat

**参数：**
- `id`: 对话ID（用于维护对话历史）
- `message`: 用户消息内容

**响应：**
- 流式返回AI回复内容（Server-Sent Events格式）

### GET /health

**响应：**
```json
{
  "status": "ok",
  "service": "eino-minimal"
}
```

## 集成到其他项目

### 1. 复制文件

将 `eino-minimal` 目录复制到你的项目中。

### 2. 修改模块名

在 `go.mod` 中修改模块名：
```go
module your-project-name/eino-minimal
```

### 3. 更新导入路径

在所有 `.go` 文件中更新导入路径：
```go
import "your-project-name/eino-minimal/agent"
```

### 4. 集成到现有HTTP服务

如果你已有HTTP服务，可以直接使用 `api.ChatHandler()` 函数：

```go
import "your-project-name/eino-minimal/api"

// 在你的路由设置中
router.GET("/chat", api.ChatHandler())
```

### 5. 直接使用代理功能

如果只需要AI对话功能，不需要HTTP接口：

```go
import "your-project-name/eino-minimal/agent"

// 直接调用代理
sr, err := agent.RunAgent(ctx, "conversation-id", "用户消息")
```

## 自定义配置

### 修改系统提示词

编辑 `agent/chat_template.go` 中的 `systemPrompt` 变量。

### 添加工具

在 `agent/tools.go` 中的 `GetTools()` 函数中添加自定义工具。

### 修改模型选择逻辑

编辑 `models/chat_model.go` 中的 `GetDefaultChatModel()` 函数。

### 调整内存配置

编辑 `memory/simple_memory.go` 中的 `GetDefaultMemory()` 函数。

## 注意事项

1. 确保设置了正确的API密钥和模型配置
2. 对话历史存储在 `./data/memory` 目录下
3. 默认对话窗口大小为6条消息
4. 服务默认运行在8080端口

## 依赖说明

- `github.com/cloudwego/eino`: eino框架核心
- `github.com/cloudwego/eino-ext`: eino扩展组件
- `github.com/cloudwego/hertz`: HTTP服务框架
- `github.com/hertz-contrib/sse`: 服务器发送事件支持
