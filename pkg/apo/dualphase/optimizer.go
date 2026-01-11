package dualphase

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"
)

// NewOptimizer 创建优化器
func NewOptimizer(g *genkit.Genkit, opts Options) *Optimizer {
	opts.Complete()
	return &Optimizer{
		opts:       opts,
		initFlow:   DefineInitializationFlow(g),
		divideFlow: DefineDivideToSentencesFlow(g),
		evalFlow:   DefineBatchEvaluationFlow(g),
		optFlow:    DefineOptimizationFlow(g),
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
	optFlow    *core.Flow[OptimizationInput, OptimizationOutput, struct{}]

	prompts        []string
	curPrompt      PromptSentences
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
		o.curAccuracy, o.curFailedCases, err = o.evaluate(ctx, o.curPrompt.String(), false)
		if err != nil {
			return nil, 0, fmt.Errorf("evaluate error: %w", err)
		}
	}

	return o.curPrompt.Copy(), o.curAccuracy, nil
}

// initP0 初始化待优化 Prompt P0
func (o *Optimizer) initP0(ctx context.Context) error {
	if o.opts.Initialization.ContinueWith != nil {
		o.curPrompt = o.opts.Initialization.ContinueWith.Copy()
		o.prompts = []string{o.curPrompt.String()}
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
		if strings.HasPrefix(s.Content, "###") && strings.HasSuffix(s.Content, "###") {
			o.curPrompt = append(o.curPrompt, WeightedSentence{
				Sentence: s,
				Ignore:   true,
			})
		} else {
			o.curPrompt = append(o.curPrompt, WeightedSentence{
				Sentence: s,
				Weight:   1,
			})
		}
	}

	o.prompts = []string{o.curPrompt.String()}

	return nil
}

// evaluate 评估
func (o *Optimizer) evaluate(ctx context.Context, prompt string, failedOnly bool) (float64, []FailedCase, error) {
	data := o.opts.Evaluation.ValidationData
	if failedOnly {
		data = make([]InputOutputPair, len(o.curFailedCases))
		for i, c := range o.curFailedCases {
			data[i] = InputOutputPair{
				Input:  c.Input,
				Output: c.Expected,
			}
		}
	}

	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})

	// 确定验证数据量和批次大小
	dataLen := len(data)
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
		if endI > dataLen {
			endI = dataLen
		}

		out, err := o.evalFlow.Run(ctx, BatchEvaluationInput{
			Prompt:         prompt,
			ValidationData: data[i*batchSize : endI],
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
func (o *Optimizer) Optimize(ctx context.Context) (PromptSentences, float64, error) {
	logger := logr.FromContextOrDiscard(ctx)

	if len(o.curFailedCases) == 0 {
		return o.curPrompt.Copy(), o.curAccuracy, fmt.Errorf("current result are the best")
	}

	// 选择待优化句子
	sentenceI, sentence := o.curPrompt.Sample()
	newPrompt := o.curPrompt.Copy()
	var undesiredSentences []string
	for i := 0; i < 6; i++ {
		// 优化
		out, err := o.optFlow.Run(ctx, OptimizationInput{
			Prompt:      o.curPrompt.String(),
			Sentence:    sentence.Content,
			FailedCases: o.curFailedCases,
		})
		if err != nil {
			return nil, 0, err
		}
		newPrompt[sentenceI].Content = out.NewSentence

		if slices.Contains(undesiredSentences, out.NewSentence) {
			// 重复输出，重试，这次重试不算次数
			i--
			continue
		}

		// 新句子在失败集上快速评估
		// 参考文章中 Eq. 7
		accuracyF, _, err := o.evaluate(ctx, newPrompt.String(), true)
		if err != nil {
			return nil, 0, fmt.Errorf("evaluate error: %w", err)
		}
		logger.Info(fmt.Sprintf(
			"new prompt's accuracy in failed cases: %.4f (expected >%.2f)",
			accuracyF, o.opts.Optimization.Hf,
		))
		if accuracyF < o.opts.Optimization.Hf {
			if i > 2 {
				// 多次重试仍不理想，重新选择句子优化
				// 这不符合文章，但是感觉更合理
				for {
					newChoiceI, newChoiceSentence := o.curPrompt.Sample()
					if len(o.curPrompt) == 1 || newChoiceI != sentenceI {
						sentenceI, sentence = newChoiceI, newChoiceSentence
						break
					}
				}
				newPrompt = o.curPrompt.Copy()
				undesiredSentences = nil
				continue
			}
			// 未通过快速评估，重新优化该句子
			undesiredSentences = append(undesiredSentences, out.NewSentence)
			continue
		}

		// 完整评估
		accuracyV, failedCases, err := o.evaluate(ctx, newPrompt.String(), false)
		if err != nil {
			return nil, 0, fmt.Errorf("evaluate error: %w", err)
		}
		logger.Info(fmt.Sprintf(
			"new prompt's accuracy in full cases: %.4f (old: %.4f, expected >+%.2f)",
			accuracyV, o.curAccuracy, o.opts.Optimization.Hv,
		))

		// 检查新 Prompt 效果提升是否达到阈值
		// 参考文章中 Eq. 8
		if accuracyV <= o.curAccuracy ||
			(o.curAccuracy < 1-o.opts.Optimization.Hv && accuracyV-o.curAccuracy < o.opts.Optimization.Hv) {
			// 提升效果不及预期，重新选择句子优化
			for {
				newChoiceI, newChoiceSentence := o.curPrompt.Sample()
				if len(o.curPrompt) == 1 || newChoiceI != sentenceI {
					sentenceI, sentence = newChoiceI, newChoiceSentence
					break
				}
			}
			newPrompt = o.curPrompt.Copy()
			undesiredSentences = nil
			continue
		}

		// 新 Prompt 在验证集和失败集上的综合效果
		accuracyMixed := accuracyV*o.opts.Optimization.MixingRate + (1-o.opts.Optimization.MixingRate)*accuracyF
		// 更新句子权重
		newPrompt.UpdateWeight(accuracyMixed, o.opts.Optimization.LearningRate)

		weights := make([]float64, len(newPrompt))
		for j, p := range newPrompt {
			weights[j] = p.Weight
		}
		logger.Info(fmt.Sprintf("Weights: %v", weights))

		o.prompts = append(o.prompts, newPrompt.String())
		o.curAccuracy = accuracyV
		o.curFailedCases = failedCases
		o.curPrompt = newPrompt

		return o.curPrompt.Copy(), o.curAccuracy, nil
	}

	return o.curPrompt.Copy(), o.curAccuracy, fmt.Errorf("current result have already converged")
}
