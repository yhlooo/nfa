package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/firebase/genkit/go/core/api"
	oai "github.com/firebase/genkit/go/plugins/compat_oai"
)

const (
	DeepseekProviderName = "deepseek"
	DeepseekBaseURL      = "https://api.deepseek.com"
)

// DeepseekOptions Deepseek 选项
type DeepseekOptions struct {
	// Deepseek API 密钥
	APIKey string `json:"apiKey"`
}

// DeepseekPlugin 基于选项创建 Deepseek 插件
func (opts *DeepseekOptions) DeepseekPlugin() *oai.OpenAICompatible {
	return &oai.OpenAICompatible{
		Provider: DeepseekProviderName,
		BaseURL:  DeepseekBaseURL,
		APIKey:   opts.APIKey,
	}
}

// ListOpenAICompatibleModelsResponse 列出 OpenAI 模型响应
type ListOpenAICompatibleModelsResponse struct {
	Data []OpenAICompatibleModel `json:"data,omitempty"`
}

// OpenAICompatibleModel OpenAI 兼容的模型
type OpenAICompatibleModel struct {
	ID string `json:"id"`
}

// ListOpenAICompatibleModels 列出 OpenAI 兼容的模型
func ListOpenAICompatibleModels(ctx context.Context, plugin *oai.OpenAICompatible) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, plugin.BaseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("make request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if plugin.APIKey != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", plugin.APIKey))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response body error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d (!=200), body: %s", resp.StatusCode, string(body))
	}

	respData := &ListOpenAICompatibleModelsResponse{}
	if err := json.Unmarshal(body, respData); err != nil {
		return nil, fmt.Errorf("unmarshal response body error: %w, body: %s", err, string(body))
	}

	if len(respData.Data) == 0 {
		return nil, nil
	}

	ret := make([]string, 0, len(respData.Data))
	for _, item := range respData.Data {
		if item.ID == "" {
			continue
		}
		ret = append(ret, api.NewName(plugin.Provider, item.ID))
	}

	return ret, nil
}
