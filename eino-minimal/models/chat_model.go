package models

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
)

// NewArkChatModel 创建豆包聊天模型
func NewArkChatModel(ctx context.Context) model.ToolCallingChatModel {
	chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		Model:   os.Getenv("ARK_MODEL"),
		APIKey:  os.Getenv("ARK_API_KEY"),
		BaseURL: os.Getenv("ARK_BASE_URL"),
	})
	if err != nil {
		log.Printf("豆包模型创建失败: %v", err)
		return nil
	}
	return chatModel
}

// NewQwenChatModel 创建千问聊天模型
func NewQwenChatModel(ctx context.Context) model.ToolCallingChatModel {
	chatModel, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		Model:   os.Getenv("QWEN_MODEL"),
		APIKey:  os.Getenv("QWEN_API_KEY"),
		BaseURL: os.Getenv("QWEN_BASE_URL"),
	})
	if err != nil {
		log.Printf("千问模型创建失败: %v", err)
		return nil
	}
	return chatModel
}

// NewOllamaModel 创建Ollama本地模型
func NewOllamaModel(ctx context.Context) model.ToolCallingChatModel {
	chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434", // Ollama 服务地址
		Model:   "qwen3:latest",           // 模型名称
	})
	if err != nil {
		log.Printf("Ollama模型创建失败: %v", err)
		return nil
	}
	return chatModel
}

// GetDefaultChatModel 获取默认聊天模型（优先使用环境变量配置）
func GetDefaultChatModel(ctx context.Context) model.ToolCallingChatModel {
	// 优先尝试千问模型
	if os.Getenv("QWEN_API_KEY") != "" {
		return NewQwenChatModel(ctx)
	}
	
	// 其次尝试豆包模型
	if os.Getenv("ARK_API_KEY") != "" {
		return NewArkChatModel(ctx)
	}
	
	// 最后尝试本地Ollama模型
	return NewOllamaModel(ctx)
}
