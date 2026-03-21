package models

// Models 模型配置
type Models struct {
	// 主模型，用于回答用户问题
	Primary string `json:"primary,omitempty"`
	// 轻量模型，用于处理简单事务。为空时使用主模型
	Light string `json:"light,omitempty"`
	// 视觉模型，用于处理图片理解任务
	Vision string `json:"vision,omitempty"`

	// 所有模型信息
	allModels map[string]ModelConfig
}

// SetAllModels 设置所有模型信息
func (m *Models) SetAllModels(models []ModelConfig) {
	if m.allModels == nil {
		m.allModels = make(map[string]ModelConfig, len(models))
	}
	for _, model := range models {
		m.allModels[model.Name] = model
	}
}

// GetPrimary 获取主模型
func (m *Models) GetPrimary() string {
	return m.Primary
}

// GetLight 获取轻量模型
func (m *Models) GetLight() string {
	if m.Light != "" {
		return m.Light
	}
	return m.Primary
}

// GetVision 获取视觉模型
func (m *Models) GetVision() string {
	if m.Vision != "" {
		return m.Vision
	}
	if m.allModels[m.Primary].Vision {
		return m.Primary
	}
	return ""
}
