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
	Channels ChannelsConfig `json:"channels,omitempty"`
	// 语言，可选 en, zh
	Language string `json:"language,omitempty"`
	// 最大上下文窗口
	// 默认 200K
	MaxContextWindow int64 `json:"maxContextWindow,omitempty"`
}

// ChannelsConfig 消息通道配置
type ChannelsConfig struct {
	Enabled  bool      `json:"enabled"`
	Channels []Channel `json:"channels,omitempty"`
}

// Channel 信道
type Channel struct {
	WeComAIBot *WeComAIBotOptions `json:"wecomAIBot,omitempty"`
	YuanbaoBot *YuanbaoBotOptions `json:"yuanbaoBot,omitempty"`
}

// WeComAIBotOptions 企业微信智能机器人选项
type WeComAIBotOptions struct {
	BotID  string `json:"botID"`
	Secret string `json:"secret"`
	URL    string `json:"url,omitempty"`
}

// YuanbaoBotOptions 元宝机器人选项
type YuanbaoBotOptions struct {
	AppID        string `json:"appID"`
	AppSecret    string `json:"appSecret"`
	BaseURL      string `json:"baseURL,omitempty"`
	WebSocketURL string `json:"websocketURL,omitempty"`
}
