package configs

import (
	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/models"
)

// Config 配置
type Config struct {
	// 模型提供商配置
	ModelProviders []models.ModelProvider `json:"modelProviders,omitempty"`
	// 默认模型
	DefaultModels models.Models `json:"defaultModels,omitempty"`
	// 数据提供商配置
	DataProviders []agents.DataProvider `json:"dataProviders,omitempty"`
	// 语言，可选 en, zh
	Language string `json:"language,omitempty"`
}
