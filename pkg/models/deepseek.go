package models

const (
	// DeepseekProviderName Deepseek 模型供应商名
	DeepseekProviderName = "deepseek"
	// DeepseekBaseURL Deepseek 默认 API 地址
	DeepseekBaseURL = "https://api.deepseek.com"
)

var (
	DeepSeekV4Pro = ModelConfig{
		Name:      "deepseek-v4-pro",
		Reasoning: true,
		Prices: ModelPrices{
			Input:  12,
			Output: 24,
			Cached: 1,
		},
		ContextWindow: 1000000,
		Score:         10,
	}
	DeepSeekV4Flash = ModelConfig{
		Name:      "deepseek-v4-flash",
		Reasoning: true,
		Prices: ModelPrices{
			Input:  1,
			Output: 2,
			Cached: 0.2,
		},
		ContextWindow: 1000000,
		Score:         8,
	}
)

// DeepSeekModels 建议的 DeepSeek 模型
var DeepSeekModels = []ModelConfig{
	DeepSeekV4Pro,
	DeepSeekV4Flash,
}
