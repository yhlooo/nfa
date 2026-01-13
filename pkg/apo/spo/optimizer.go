package spo

import (
	"context"
	"fmt"
	"io"
	"strings"

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

	OutputWriter io.Writer `json:"-"`
}

// NewOptimizer 创建优化器
func NewOptimizer(g *genkit.Genkit, opts Options) *Optimizer {
	w := opts.OutputWriter
	if w == nil {
		w = io.Discard
	}
	return &Optimizer{
		opts:     opts,
		optFlow:  DefineOptimizationFlow(g),
		exeFlow:  DefineExecutionFlow(g),
		evalFlow: DefineEvaluationFlow(g),
		w:        w,
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

	w io.Writer

	promptI    int
	draftI     int
	curPrompt  string
	curAnswers []string
}

// Initialize 初始化
func (o *Optimizer) Initialize(ctx context.Context) error {
	o.curPrompt = o.opts.P0
	o.writePrompt(o.promptI, "", o.curPrompt)

	answers := make([]string, len(o.opts.EvaluationQuestions))
	for i, q := range o.opts.EvaluationQuestions {
		o.writeAnswerTitle(i + 1)
		exeRet, err := o.exeFlow.Run(ctx, ExecutionInput{
			Prompt:    o.curPrompt,
			Questions: []string{q},
		})
		if err != nil {
			return fmt.Errorf("execute prompt error: %w", err)
		}
		_, _ = fmt.Fprint(o.w, "\n\n")
		if len(exeRet.Answers) > 0 {
			answers[i] = exeRet.Answers[0]
		}
	}

	o.curAnswers = answers
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
	o.writeOptimizationResult(o.promptI+1, o.draftI, optRet)

	// 执行新 Prompt

	oldBetter := 0
	newBetter := 0
	newAnswers := make([]string, len(o.opts.EvaluationQuestions))
	for i, q := range o.opts.EvaluationQuestions {
		o.writeAnswerTitle(i + 1)
		exeRet, err := o.exeFlow.Run(ctx, ExecutionInput{
			Prompt:    optRet.Prompt,
			Questions: []string{q},
		})
		if err != nil {
			return optRet.Prompt, false, fmt.Errorf("execute new prompt error: %w", err)
		}
		_, _ = fmt.Fprint(o.w, "\n\n")
		if len(exeRet.Answers) > 0 {
			newAnswers[i] = exeRet.Answers[0]
		}

		o.writeEvalTitle()
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
				return optRet.Prompt, false, fmt.Errorf("evaluation error: %w", err)
			}
			o.writeEvalResult(j%2 != 0, evalRet)

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
	o.writeEvalFinalResult(oldBetter, newBetter)

	// 如果新的更好，替换旧的 Prompt
	if newBetter > oldBetter {
		o.curAnswers = newAnswers
		o.curPrompt = optRet.Prompt
		o.promptI++
		o.draftI = 0
		o.writePrompt(o.promptI, "Accepted", o.curPrompt)
	} else {
		o.draftI++
	}

	return optRet.Prompt, newBetter > oldBetter, nil
}

// writeOptimizationResult 写优化结果
func (o *Optimizer) writeOptimizationResult(promptI, draftI int, ret OptimizationOutput) {
	_, _ = fmt.Fprintf(o.w,
		`### P%d (Draft%d)

> %s

%s

`,
		promptI,
		draftI,
		strings.ReplaceAll(strings.TrimSpace(ret.Analysis), "\n", "\n> "),
		strings.TrimSpace(ret.Prompt),
	)
}

// writePrompt 输出 Prompt
func (o *Optimizer) writePrompt(i int, tag, content string) {
	title := fmt.Sprintf("### P%d", i)
	if tag != "" {
		title += " (" + tag + ")"
	}

	_, _ = fmt.Fprintf(o.w, `%s

%s

`, title, strings.TrimSpace(content))
}

// writeAnswerTitle 输出答案
func (o *Optimizer) writeAnswerTitle(i int) {
	_, _ = fmt.Fprintf(o.w, "#### Answer %d\n\n", i)
}

// writeEvalTitle 写评估标题
func (o *Optimizer) writeEvalTitle() {
	_, _ = fmt.Fprint(o.w, "#### Evaluation\n\n")
}

// writeEvalResult 写评估结果
func (o *Optimizer) writeEvalResult(aIsCurrent bool, ret EvaluationOutput) {
	nameLine := "**A -> New** | B -> Current"
	choices := map[string]string{
		"A": "New",
		"B": "Current",
	}
	if aIsCurrent {
		nameLine = "A -> Current | **B -> New**"
		choices = map[string]string{
			"A": "Current",
			"B": "New",
		}
	}

	_, _ = fmt.Fprintf(o.w, `---

%s

> %s

Better: %s

`,
		nameLine,
		strings.ReplaceAll(strings.TrimSpace(ret.Analysis), "\n", "\n> "),
		choices[ret.Choice],
	)
}

// writeEvalFinalResult 写结果行
func (o *Optimizer) writeEvalFinalResult(currentBetter, newBetter int) {
	better := "Current"
	if newBetter > currentBetter {
		better = "New"
	}
	_, _ = fmt.Fprintf(o.w, `#### Evaluation Summary

**Final Result: %d:%d (%s Better)**

`, currentBetter, newBetter, better)
}
