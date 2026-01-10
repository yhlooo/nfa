package dualphase

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// NewOptimizer 创建优化器
func NewOptimizer(g *genkit.Genkit, opts Options) *Optimizer {
	return &Optimizer{
		opts:       opts,
		initFlow:   DefineInitializationFlow(g),
		divideFlow: DefineDivideToSentencesFlow(g),
		evalFlow:   DefineBatchEvaluationFlow(g),
	}
}

// Optimizer 双阶段加速的 Prompt 优化器
//
// 参考 https://arxiv.org/abs/2406.13443 (Dual-Phase Accelerated Prompt Optimization)
type Optimizer struct {
	opts       Options
	initFlow   *core.Flow[InitializationInput, InitializationOutput, struct{}]
	divideFlow *core.Flow[DivideToSentencesInput, DivideToSentencesOutput, struct{}]
	evalFlow   *core.Flow[BatchEvaluationInput, BatchEvaluationOutput, struct{}]

	prompts        []string
	curSentences   PromptSentences
	curAccuracy    float64
	curFailedCases []FailedCase
}

// Initialize 初始化
//
// 生成初始待优化 Prompt P0
func (o *Optimizer) Initialize(ctx context.Context) (PromptSentences, float64, error) {
	if err := o.initP0(ctx); err != nil {
		return nil, 0, err
	}

	// 进行首轮评估
	var err error
	o.curAccuracy = 1
	if len(o.opts.Evaluation.ValidationData) != 0 {
		o.curAccuracy, o.curFailedCases, err = o.evaluate(ctx, o.curSentences.String())
		if err != nil {
			return nil, 0, fmt.Errorf("evaluate error: %w", err)
		}
	}

	return o.curSentences.Copy(), o.curAccuracy, nil
}

// initP0 初始化待优化 Prompt P0
func (o *Optimizer) initP0(ctx context.Context) error {
	if o.opts.Initialization.ContinueWith != nil {
		o.curSentences = o.opts.Initialization.ContinueWith.Copy()
		o.prompts = []string{o.curSentences.String()}
		return nil
	}

	// 生成 p0
	p0 := o.opts.Initialization.P0
	if p0 == "" {
		initOut, err := o.initFlow.Run(ctx, InitializationInput{
			PreviousP0:   o.opts.Initialization.PreviousP0,
			TrainingData: o.opts.Initialization.TrainingData,
			InChinese:    o.opts.Initialization.InChinese,
		})
		if err != nil {
			return fmt.Errorf("generate prompt p0 error: %w", err)
		}
		p0 = initOut.Prompt
	}

	// 将 p0 按句子切分
	divideOut, err := o.divideFlow.Run(ctx, DivideToSentencesInput{Content: p0})
	if err != nil {
		return fmt.Errorf("divide prompt p0 to sentences error: %w", err)
	}

	// 为每个句子赋予初始权重 1
	for _, s := range divideOut.Sentences {
		o.curSentences = append(o.curSentences, WeightedSentence{
			Sentence: s,
			Weight:   1,
			Ignore:   false,
		})
	}

	o.prompts = []string{o.curSentences.String()}

	return nil
}

// evaluate 评估
func (o *Optimizer) evaluate(ctx context.Context, prompt string) (float64, []FailedCase, error) {
	// 确定验证数据量和批次大小
	dataLen := len(o.opts.Evaluation.ValidationData)
	if dataLen == 0 {
		return 0, nil, fmt.Errorf("no validation data")
	}
	batchSize := o.opts.Evaluation.EvaluationBatchSize
	if batchSize <= 0 {
		batchSize = 1
	}

	// 分批评估
	correct := 0
	wrong := 0
	var failedCases []FailedCase
	for i := 0; i*batchSize < dataLen; i++ {
		endI := (i + 1) * batchSize
		if endI >= dataLen {
			endI = dataLen - 1
		}

		out, err := o.evalFlow.Run(ctx, BatchEvaluationInput{
			Prompt:         prompt,
			ValidationData: o.opts.Evaluation.ValidationData[i*batchSize : endI],
		})
		if err != nil {
			return float64(correct) / float64(correct+wrong), failedCases, err
		}
		correct += out.Correct
		wrong += out.Wrong
		failedCases = append(failedCases, out.FailedCases...)
	}

	return float64(correct) / float64(correct+wrong), failedCases, nil
}

// Optimize 进行一轮优化
func (o *Optimizer) Optimize(ctx context.Context) (PromptSentences, error) {
	return nil, nil
}
