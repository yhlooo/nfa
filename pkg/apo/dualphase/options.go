package dualphase

// Options 选项
type Options struct {
	// 初始化选项
	Initialization InitializationOptions `json:"initialization"`
}

// InitializationOptions 初始化选项
type InitializationOptions struct {
	// 初始 Prompt
	// 若值不为空则跳过初始化步骤
	P0 string `json:"p0,omitempty"`

	// 初始 Prompt 前的 Prompt
	PreviousP0 string `json:"previousP0,omitempty"`
	// 使用 PreviousP0 生成的训练数据数量
	GenerateTrainingDataPairs int `json:"generateTrainingDataPairs,omitempty"`
	// 用于生成初始 Prompt 的训练数据
	TrainingData []TrainingData `json:"trainingData,omitempty"`

	// 是否生成中文 P0
	InChinese bool `json:"inChinese,omitempty"`
}
