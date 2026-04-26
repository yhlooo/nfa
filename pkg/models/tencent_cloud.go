package models

const (
	// TencentProviderName 腾讯云 TokenHub 模型供应商名
	TencentProviderName = "tencent"
	// TencentBaseURL 腾讯云 TokenHub 默认 API 地址
	TencentBaseURL = "https://tokenhub.tencentmaas.com/v1"
)

var (
	HY3Preview = ModelConfig{
		Name:      "hy3-preview",
		Reasoning: true,
		Prices: ModelPrices{
			Input:  2, // > 32K
			Output: 8,
			Cached: 0.8,
		},
		ContextWindow: 256000,
		Score:         3,
	}
)

// TencentModels 腾讯云推荐模型
var TencentModels = []ModelConfig{
	HY3Preview,
}
