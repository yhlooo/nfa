package dualphase

// Options 选项
type Options struct {
	// 初始化选项
	Initialization InitializationOptions `json:"initialization,omitempty"`
	// 评估选项
	Evaluation EvaluationOptions `json:"evaluation,omitempty"`
	// 优化选项
	Optimization OptimizationOptions `json:"optimization,omitempty"`
}

// Complete 使用默认值补充选项
func (opts *Options) Complete() {
	opts.Evaluation.Complete()
	opts.Optimization.Complete()
}

// InitializationOptions 初始化选项
type InitializationOptions struct {
	// 初始 Prompt
	// 若值不为空则跳过初始化步骤
	P0 string `json:"p0,omitempty"`
	// 从指定中间结果继续优化
	ContinueWith PromptSentences `json:"continueWith,omitempty"`

	// 初始 Prompt 前的 Prompt
	PreviousP0 string `json:"previousP0,omitempty"`
	// 使用 PreviousP0 生成的训练数据数量
	GenerateTrainingDataPairs int `json:"generateTrainingDataPairs,omitempty"`
	// 用于生成初始 Prompt 的训练数据
	TrainingData []InputOutputPair `json:"trainingData,omitempty"`

	// 是否生成中文 P0
	InChinese bool `json:"inChinese,omitempty"`
}

// EvaluationOptions 评估选项
type EvaluationOptions struct {
	// 用于验证 Prompt 的数据
	ValidationData []InputOutputPair `json:"validationData,omitempty"`
	// 评估批次大小（每轮生成验证多少对输入输出）
	EvaluationBatchSize int `json:"evaluationBatchSize,omitempty"`
}

// Complete 使用默认值补充选项
func (opts *EvaluationOptions) Complete() {
	if opts.EvaluationBatchSize == 0 {
		opts.EvaluationBatchSize = 1
	}
}

// OptimizationOptions 优化选项
type OptimizationOptions struct {
	// 每轮迭代在上轮失败集上的提升幅度阈值
	Hf float64 `json:"hf,omitempty"`
	// 每轮迭代在验证集上的提升幅度阈值
	Hv float64 `json:"hv,omitempty"`
	// 综合效果中混合验证集效果的比例（文章中 Eq. 9 中的 α ）
	MixingRate float64 `json:"alpha,omitempty"`
	// 学习率（文章中 Eq. 9 中的 μ ）
	LearningRate float64 `json:"learningRate,omitempty"`
}

// Complete 使用默认值补充选项
func (opts *OptimizationOptions) Complete() {
	if opts.Hf == 0 {
		opts.Hf = 0.3
	}
	if opts.Hv == 0 {
		opts.Hv = 0.1
	}
	if opts.MixingRate == 0 {
		opts.MixingRate = 0.4
	}
	if opts.LearningRate == 0 {
		opts.LearningRate = 0.055
	}
}

// InputOutputPair 输入输出对
type InputOutputPair struct {
	// 输入
	Input string `json:"input"`
	// 输出
	Output string `json:"output"`
}
