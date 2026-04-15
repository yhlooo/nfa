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
	DataProviders agents.DataProviders `json:"dataProviders,omitempty"`
	// 信道
	Channels []Channel `json:"channels,omitempty"`
	// 语言，可选 en, zh
	Language string `json:"language,omitempty"`
}

// Channel 信道
type Channel struct {
	WeComAIBot *WeComAIBotOptions `json:"wecomAIBot,omitempty"`
}

// WeComAIBotOptions 企业微信智能机器人选项
type WeComAIBotOptions struct {
	BotID  string `json:"botID"`
	Secret string `json:"secret"`
	URL    string `json:"url,omitempty"`
}
