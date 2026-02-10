package models

// ModelProvider 模型供应商配置
type ModelProvider struct {
	Ollama           *OllamaOptions           `json:"ollama,omitempty"`
	Deepseek         *DeepseekOptions         `json:"deepseek,omitempty"`
	OpenAICompatible *OpenAICompatibleOptions `json:"openaiCompatible,omitempty"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	// 模型名称（必需）
	Name string `json:"name"`

	// 是否支持推理/思考模式（预留）
	Reasoning bool `json:"reasoning,omitempty"`

	// 是否支持视觉/图片理解（预留）
	Vision bool `json:"vision,omitempty"`

	// 价格信息（预留），单位：元/千Token
	Cost ModelCost `json:"cost,omitempty"`

	// 上下文窗口大小（预留）
	ContextWindow int `json:"contextWindow,omitempty"`

	// 最大输出 Token 数（预留）
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
}

// ModelCost 价格信息
type ModelCost struct {
	Input  float64 `json:"input,omitempty"`
	Output float64 `json:"output,omitempty"`
}
