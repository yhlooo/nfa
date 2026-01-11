package spo

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"
)

// Options 优化器选项
type Options struct {
	// 待优化的 Prompt
	P0 string `json:"p0"`
	// 额外要求
	Requirements string `json:"requirements,omitempty"`
	// 用于评估的问题
	EvaluationQuestions []string `json:"evaluationQuestions"`
}

// NewOptimizer 创建优化器
func NewOptimizer(g *genkit.Genkit, opts Options) *Optimizer {
	return &Optimizer{
		opts:     opts,
		optFlow:  DefineOptimizationFlow(g),
		exeFlow:  DefineExecutionFlow(g),
		evalFlow: DefineEvaluationFlow(g),
	}
}

// Optimizer 自监督 Prompt 优化器
//
// 参考 https://arxiv.org/abs/2502.06855 (Self-Supervised Prompt Optimization)
type Optimizer struct {
	opts     Options
	optFlow  *core.Flow[OptimizationInput, OptimizationOutput, struct{}]
	exeFlow  *core.Flow[ExecutionInput, ExecutionOutput, struct{}]
	evalFlow *core.Flow[EvaluationInput, EvaluationOutput, struct{}]

	curPrompt  string
	curAnswers []string
}

// Initialize 初始化
func (o *Optimizer) Initialize(ctx context.Context) error {
	o.curPrompt = o.opts.P0
	var err error

	exeRet, err := o.exeFlow.Run(ctx, ExecutionInput{
		Prompt:    o.curPrompt,
		Questions: o.opts.EvaluationQuestions,
	})
	if err != nil {
		return fmt.Errorf("execute p0 error: %w", err)
	}

	o.curAnswers = exeRet.Answers
	return nil
}

// Optimize 进行一轮优化
func (o *Optimizer) Optimize(ctx context.Context) (string, bool, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// 优化产生新 Prompt
	optRet, err := o.optFlow.Run(ctx, OptimizationInput{
		Requirements: o.opts.Requirements,
		Prompt:       o.curPrompt,
		Answers:      o.curAnswers,
	})
	if err != nil {
		return "", false, fmt.Errorf("optimization error: %w", err)
	}

	// 执行新 Prompt
	newPrompt := optRet.Prompt
	exeRet, err := o.exeFlow.Run(ctx, ExecutionInput{
		Prompt:    newPrompt,
		Questions: o.opts.EvaluationQuestions,
	})
	if err != nil {
		return newPrompt, false, fmt.Errorf("execute new prompt error: %w", err)
	}
	newAnswers := exeRet.Answers
	if len(newAnswers) != len(o.curAnswers) {
		return newPrompt, false, fmt.Errorf(
			"the number of new answers not match: %d (expected %d)",
			len(newAnswers), len(o.curAnswers),
		)
	}

	oldBetter := 0
	newBetter := 0
	for i := range newAnswers {
		for j := 0; j < 4; j++ {
			// 交换两个答案评估 4 次
			var a, b string
			if j%2 == 0 {
				a = newAnswers[i]
				b = o.curAnswers[i]
			} else {
				a = o.curAnswers[i]
				b = newAnswers[i]
			}

			evalRet, err := o.evalFlow.Run(ctx, EvaluationInput{
				Requirements: o.opts.Requirements,
				AnswerA:      a,
				AnswerB:      b,
			})
			if err != nil {
				return newPrompt, false, fmt.Errorf("evaluation error: %w", err)
			}

			switch evalRet.Choice {
			case "A":
				if j%2 == 0 {
					newBetter++
				} else {
					oldBetter++
				}
			case "B":
				if j%2 == 0 {
					oldBetter++
				} else {
					newBetter++
				}
			}
		}
	}
	logger.Info(fmt.Sprintf("old:new -> %d:%d", oldBetter, newBetter))

	// 如果新的更好，替换旧的 Prompt
	if newBetter > oldBetter {
		o.curAnswers = newAnswers
		o.curPrompt = newPrompt
	}

	return newPrompt, newBetter > oldBetter, nil
}
