package configs

import "github.com/yhlooo/nfa/pkg/agents/models"

// Config 配置
type Config struct {
	// 模型供应商配置
	ModelProviders []models.ModelProvider `json:"modelProviders,omitempty"`
}
