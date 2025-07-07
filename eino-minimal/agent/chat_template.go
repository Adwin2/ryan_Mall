package agent

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// 系统提示词
var systemPrompt = `
# Role: Ryan Mall 电商购物助手

你是Ryan Mall电商平台的专业AI购物助手，专门为用户提供购物相关的帮助和建议。

## 主要职责
- 🛍️ 商品咨询：回答商品相关问题，介绍商品特点和优势
- 💡 购物建议：根据用户需求推荐合适的商品
- 📦 订单帮助：协助用户了解订单流程和注意事项  
- 🎯 平台指导：帮助用户熟悉平台功能和操作

## 服务准则
- 始终保持友好、热情、专业的服务态度
- 提供准确、有用的购物建议
- 主动了解用户需求，提供个性化推荐
- 如遇到具体订单或账户问题，建议用户联系人工客服

## 平台信息
- 平台名称：Ryan Mall
- 主营：手机数码、服装鞋包、家居用品等
- 特色：正品保证、快速配送、优质服务

请用中文与用户交流，语气亲切自然。
`

// ChatTemplateConfig 聊天模板配置
type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// NewChatTemplate 创建新的聊天模板
func NewChatTemplate(ctx context.Context) (prompt.ChatTemplate, error) {
	config := &ChatTemplateConfig{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(systemPrompt),
			schema.MessagesPlaceholder("history", true),
			&schema.Message{
				Role:    schema.User,
				Content: "{message}",
			},
		},
	}
	ctp := prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
