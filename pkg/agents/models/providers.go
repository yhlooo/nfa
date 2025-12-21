package models

// ModelProvider 模型供应商配置
type ModelProvider struct {
	Ollama   *OllamaOptions   `json:"ollama,omitempty"`
	Deepseek *DeepseekOptions `json:"deepseek,omitempty"`
}
