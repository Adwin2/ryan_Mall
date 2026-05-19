package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cloudwego/eino/components/embedding"
)

type ArkEmbedder struct {
	apiKey  string
	baseURL string
	model   string
}

func NewArkEmbedder(model string) *ArkEmbedder {
	baseURL := os.Getenv("ARK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	return &ArkEmbedder{
		apiKey:  os.Getenv("ARK_API_KEY"),
		baseURL: baseURL,
		model:   model,
	}
}

func (e *ArkEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("ARK_API_KEY not set")
	}

	body := map[string]any{
		"model": e.model,
		"input": texts,
	}
	data, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse embedding response: %w", err)
	}

	vectors := make([][]float64, len(result.Data))
	for i, d := range result.Data {
		vectors[i] = d.Embedding
	}
	return vectors, nil
}
