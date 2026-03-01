package models

// Models 模型配置
type Models struct {
	// 主模型，用于回答用户问题
	Primary string `json:"primary,omitempty"`
	// 轻量模型，用于处理简单事务。为空时使用主模型
	Light string `json:"light,omitempty"`
	// 视觉模型，用于处理图片理解任务
	Vision string `json:"vision,omitempty"`
}

// GetPrimary 获取主模型
func (m Models) GetPrimary() string {
	return m.Primary
}

// GetLight 获取轻量模型
func (m Models) GetLight() string {
	if m.Light != "" {
		return m.Light
	}
	return m.Primary
}

// GetVision 获取视觉模型
func (m Models) GetVision() string {
	return m.Vision
}
