package models

// ModelProvider 模型供应商配置
type ModelProvider struct {
	Ollama           *OllamaOptions           `json:"ollama,omitempty"`
	Zhipu            *BigModelOptions         `json:"zhipu,omitempty"`
	Aliyun           *DashScopeOptions        `json:"aliyun,omitempty"`
	Deepseek         *DeepseekOptions         `json:"deepseek,omitempty"`
	OpenAICompatible *OpenAICompatibleOptions `json:"openaiCompatible,omitempty"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	// 模型名称
	Name string `json:"name"`

	// 是否支持推理、思考模式
	Reasoning bool `json:"reasoning,omitempty"`
	// 是否支持视觉、图片理解
	Vision bool `json:"vision,omitempty"`

	// 价格信息
	Cost ModelCost `json:"cost,omitempty"`
	// 上下文窗口大小
	ContextWindow int64 `json:"contextWindow,omitempty"`
	// 最大输出 Token 数
	MaxOutputTokens int64 `json:"maxOutputTokens,omitempty"`
}

// ModelCost 价格信息
type ModelCost struct {
	// 每千输入 Token 价格
	Input float64 `json:"input,omitempty"`
	// 每千输出 Token 价格
	Output float64 `json:"output,omitempty"`
	// 每千缓存 Token 价格
	Cached float64 `json:"cached,omitempty"`
}
