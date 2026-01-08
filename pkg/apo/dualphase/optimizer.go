package dualphase

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

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

// NewOptimizer 创建优化器
func NewOptimizer(g *genkit.Genkit, opts Options) *Optimizer {
	return &Optimizer{
		opts:       opts,
		initFlow:   DefineInitializationFlow(g),
		divideFlow: DefineDivideToSentencesFlow(g),
	}
}

// Optimizer 双阶段加速的 Prompt 优化器
//
// 参考 https://arxiv.org/abs/2406.13443 (Dual-Phase Accelerated Prompt Optimization)
type Optimizer struct {
	opts       Options
	initFlow   *core.Flow[InitializationInput, InitializationOutput, struct{}]
	divideFlow *core.Flow[DivideToSentencesInput, DivideToSentencesOutput, struct{}]

	prompts      []string
	curSentences PromptSentences
}

// Initialization 初始化
//
// 生成初始待优化 Prompt P0
func (o *Optimizer) Initialization(ctx context.Context) (string, error) {
	// 生成 p0
	initOut, err := o.initFlow.Run(ctx, InitializationInput{
		PreviousP0:   o.opts.Initialization.PreviousP0,
		TrainingData: o.opts.Initialization.TrainingData,
		InChinese:    o.opts.Initialization.InChinese,
	})
	if err != nil {
		return "", fmt.Errorf("generate prompt p0 error: %w", err)
	}

	// 将 p0 按句子切分
	divideOut, err := o.divideFlow.Run(ctx, DivideToSentencesInput{Content: initOut.Prompt})
	if err != nil {
		return "", fmt.Errorf("divide prompt p0 to sentences error: %w", err)
	}

	// 为每个句子赋予初始权重 1
	for _, s := range divideOut.Sentences {
		o.curSentences = append(o.curSentences, WeightedSentence{
			Sentence: s,
			Weight:   1,
			Ignore:   false,
		})
	}

	o.prompts = append(o.prompts, initOut.Prompt)
	return initOut.Prompt, nil
}
