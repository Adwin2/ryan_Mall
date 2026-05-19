package routing

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino-ext/components/model/ark"
)

type ModelFactory struct {
	mu    sync.RWMutex
	cache map[string]model.ToolCallingChatModel
}

func NewModelFactory() *ModelFactory {
	return &ModelFactory{
		cache: make(map[string]model.ToolCallingChatModel),
	}
}

func (f *ModelFactory) GetModel(ctx context.Context, profile *ProfileConfig) (model.ToolCallingChatModel, error) {
	cacheKey := profile.Model

	f.mu.RLock()
	if m, ok := f.cache[cacheKey]; ok {
		f.mu.RUnlock()
		return m, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if m, ok := f.cache[cacheKey]; ok {
		return m, nil
	}

	m, err := f.createModel(ctx, profile)
	if err != nil {
		return nil, err
	}

	f.cache[cacheKey] = m
	log.Printf("[ModelFactory] created model: %s", profile.Model)
	return m, nil
}

func (f *ModelFactory) createModel(ctx context.Context, profile *ProfileConfig) (model.ToolCallingChatModel, error) {
	apiKey := os.Getenv("ARK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ARK_API_KEY not set")
	}

	cfg := &ark.ChatModelConfig{
		Model:  profile.Model,
		APIKey: apiKey,
	}
	if baseURL := os.Getenv("ARK_BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	return ark.NewChatModel(ctx, cfg)
}
