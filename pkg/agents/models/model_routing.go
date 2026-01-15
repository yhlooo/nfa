package models

// Models 模型配置
type Models struct {
	// 主模型，用于回答用户问题
	Main string `json:"main,omitempty"`
	// 快速模型，用于处理简单事务。为空时使用主模型
	Fast string `json:"fast,omitempty"`
}

// GetMain 获取主模型
func (m Models) GetMain() string {
	return m.Main
}

// GetFast 获取快速模型
func (m Models) GetFast() string {
	if m.Fast != "" {
		return m.Fast
	}
	return m.Main
}
