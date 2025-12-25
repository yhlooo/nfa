package deepseek

import (
	"context"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/core/api"
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
func (d *Deepseek) Init(ctx context.Context) []api.Action {
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

	var opts []option.RequestOption

	client := openai.NewClient(opts...)
	d.client = &client
	d.initted = true

	models, err := d.client.Models.List(ctx)
	if err == nil {
		for _, m := range models.Data {
			registerModel := d.defineModel(m.ID, ai.ModelOptions{
				Label: m.ID,
				Supports: &ai.ModelSupports{
					Multiturn:  true,
					Tools:      true,
					SystemRole: true,
					Media:      true,
					ToolChoice: true,
				},
			}, m.ID == "deepseek-reasoner")
			d.models = append(d.models, registerModel.Name())
		}
	}

	return nil
}

// RegisterModels 返回注册的模型
func (d *Deepseek) RegisterModels() []string {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.models == nil {
		return nil
	}

	ret := make([]string, len(d.models))
	copy(ret, d.models)
	return ret
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
