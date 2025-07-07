package service

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AIService interface {
    ChatWithAI(userID uint, message string) (string, error)
}

type aiService struct {
    aiBaseURL   string
    httpClient  *http.Client
}

func NewAIService() AIService {
    return &aiService{
        aiBaseURL: "http://localhost:8083", // eino-minimal服务地址
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (s *aiService) ChatWithAI(userID uint, message string) (string, error) {
    // 构建请求URL
    chatURL := fmt.Sprintf("%s/api/chat?id=user_%d&message=%s", 
        s.aiBaseURL, userID, url.QueryEscape(message))
    
    resp, err := s.httpClient.Get(chatURL)
    if err != nil {
        return "", fmt.Errorf("AI服务连接失败: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("AI服务返回错误状态: %d", resp.StatusCode)
    }
    
    // 读取流式响应
    var result strings.Builder
    scanner := bufio.NewScanner(resp.Body)

    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "data:") {
            data := strings.TrimPrefix(line, "data:")
            if data != "" && data != "[DONE]" {
                result.WriteString(data)
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return "", fmt.Errorf("读取AI响应失败: %v", err)
    }

    return result.String(), nil
}