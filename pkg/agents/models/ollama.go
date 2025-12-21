package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/go-logr/logr"
)

// OllamaOptions Ollama 选项
type OllamaOptions struct {
	// Ollama 服务端地址
	//
	// 默认 http://localhost:11434
	ServerAddress string `json:"serverAddress,omitempty"`
	// 模型响应超时时间，秒
	//
	// 默认 300
	Timeout int `json:"timeout,omitempty"`
	// 模型名列表
	//
	// 空表示使用 Ollama 已下载的所有模型
	Models []string `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *OllamaOptions) Complete() {
	if opts.ServerAddress == "" {
		opts.ServerAddress = "http://localhost:11434"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 300
	}
}

// OllamaPlugin 基于选项创建 Ollama 插件
func (opts *OllamaOptions) OllamaPlugin() *ollama.Ollama {
	opts.Complete()
	return &ollama.Ollama{
		ServerAddress: opts.ServerAddress,
		Timeout:       opts.Timeout,
	}
}

// DefineModels 注册模型
func (opts *OllamaOptions) DefineModels(
	ctx context.Context,
	g *genkit.Genkit,
	plugin *ollama.Ollama,
) ([]string, error) {
	logger := logr.FromContextOrDiscard(ctx)

	if len(opts.Models) == 0 {
		models, err := opts.ListModels(ctx)
		if err != nil {
			return nil, fmt.Errorf("list ollama models error: %w", err)
		}
		opts.Models = make([]string, len(models))
		for i, model := range models {
			opts.Models[i] = model.Name
		}
	}

	var definedModels []string
	for _, model := range opts.Models {
		logger.Info(fmt.Sprintf("define ollama model %q", model))
		m := plugin.DefineModel(g, ollama.ModelDefinition{
			Name: model,
			Type: "chat",
		}, &ai.ModelOptions{
			Label: model,
			Supports: &ai.ModelSupports{
				Multiturn:  true,
				SystemRole: true,
				Tools:      true,
			},
		})
		definedModels = append(definedModels, m.Name())
	}

	return definedModels, nil
}

// ListOllamaTagsResponse 获取 Ollama tags 列表响应
type ListOllamaTagsResponse struct {
	Models []OllamaModel `json:"models,omitempty"`
}

// OllamaModel Ollama 模型
type OllamaModel struct {
	Name  string `json:"name"`
	Model string `json:"model"`

	ModifiedAt time.Time          `json:"modified_at,omitempty"`
	Size       int64              `json:"size,omitempty"`
	Digest     string             `json:"digest,omitempty"`
	Details    OllamaModelDetails `json:"details,omitempty"`
}

// OllamaModelDetails Ollama 模型详情
type OllamaModelDetails struct {
	ParentModel       string   `json:"parent_model,omitempty"`
	Format            string   `json:"format,omitempty"`
	Family            string   `json:"family,omitempty"`
	Families          []string `json:"families,omitempty"`
	ParameterSize     string   `json:"parameter_size,omitempty"`
	QuantizationLevel string   `json:"quantization_level,omitempty"`
}

// ListModels 列出模型
func (opts *OllamaOptions) ListModels(ctx context.Context) ([]OllamaModel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.ServerAddress+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("make request error: %w", err)
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

	respData := &ListOllamaTagsResponse{}
	if err := json.Unmarshal(body, respData); err != nil {
		return nil, fmt.Errorf("unmarshal response body error: %w, body: %s", err, string(body))
	}

	return respData.Models, nil
}
