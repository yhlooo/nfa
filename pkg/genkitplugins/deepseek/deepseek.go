package deepseek

import (
	"context"
	"fmt"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Deepseek 插件
type Deepseek struct {
	Provider string
	BaseURL  string
	APIKey   string

	lock    sync.Mutex
	initted bool

	client *openai.Client
	models []string
}

var _ api.Plugin = (*Deepseek)(nil)

const (
	ProviderName = "deepseek"
	BaseURL      = "https://api.deepseek.com"
)

// Name 返回插件名
func (d *Deepseek) Name() string {
	if d.Provider != "" {
		return d.Provider
	}
	return ProviderName
}

// Init 初始化插件
func (d *Deepseek) Init(_ context.Context) []api.Action {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.initted {
		panic("plugin already initialized")
	}

	if d.BaseURL == "" {
		d.BaseURL = BaseURL
	}
	if d.Provider == "" {
		d.Provider = ProviderName
	}

	opts := []option.RequestOption{option.WithBaseURL(d.BaseURL)}
	if d.APIKey != "" {
		opts = append(opts, option.WithAPIKey(d.APIKey))
	}

	client := openai.NewClient(opts...)
	d.client = &client
	d.initted = true

	return nil
}

// RegisterModels 注册模型
func (d *Deepseek) RegisterModels(ctx context.Context, g *genkit.Genkit) ([]string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	models, err := d.client.Models.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list models error: %w", err)
	}

	for _, m := range models.Data {
		opts := ai.ModelOptions{
			Label: m.ID,
			Supports: &ai.ModelSupports{
				Multiturn:  true,
				Tools:      true,
				SystemRole: true,
				Media:      true,
				ToolChoice: true,
			},
		}
		model := d.defineModel(m.ID, opts, m.ID == "deepseek-reasoner")
		genkit.DefineModel(g, model.Name(), &opts, model.Generate)
		d.models = append(d.models, model.Name())
	}

	if d.models == nil {
		return nil, nil
	}

	ret := make([]string, len(d.models))
	copy(ret, d.models)
	return ret, nil
}

// DefineModel 定义模型
func (d *Deepseek) DefineModel(name string, opts ai.ModelOptions, thinking bool) ai.Model {
	d.lock.Lock()
	defer d.lock.Unlock()

	if !d.initted {
		panic("plugin not initialized")
	}

	return d.defineModel(name, opts, thinking)
}

// defineModel 定义模型
func (d *Deepseek) defineModel(name string, opts ai.ModelOptions, thinking bool) ai.Model {
	return ai.NewModel(api.NewName(d.Name(), name), &opts, func(
		ctx context.Context,
		req *ai.ModelRequest,
		cb core.StreamCallback[*ai.ModelResponseChunk],
	) (*ai.ModelResponse, error) {
		generator := NewModelGenerator(d.client, name, thinking).
			WithMessages(req.Messages).
			WithTools(req.Tools)

		resp, err := generator.Generate(ctx, req, cb)
		if err != nil {
			return nil, err
		}

		return resp, nil
	})
}
