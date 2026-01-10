package dualphase

import (
	"strings"
)

// WeightedSentence 带权重的句子
type WeightedSentence struct {
	Sentence

	// 权重
	Weight float64 `json:"weight,omitempty"`
	// 忽略优化句子或部分
	Ignore bool `json:"ignore,omitempty"`
}

// PromptSentences 按句子分割的 Prompt
type PromptSentences []WeightedSentence

// String 转为文本形式
func (p PromptSentences) String() string {
	ret := &strings.Builder{}
	for _, s := range p {
		ret.WriteString(s.Content)
		ret.WriteString(s.Suffix)
	}
	return ret.String()
}

var (
	weightColors = []string{"\033[48;5;17m", "\033[48;5;18m", "\033[48;5;19m", "\033[48;5;20m", "\033[48;5;21m"}
	resetColor   = "\033[0m"
)

// WithWeightColors 转为带颜色表示权重的文本形式
func (p PromptSentences) WithWeightColors() string {
	minWeight := 0.
	maxWeight := 0.
	for _, s := range p {
		if s.Weight > maxWeight {
			maxWeight = s.Weight
		}
		if s.Weight < minWeight {
			minWeight = s.Weight
		}
	}
	step := (maxWeight - minWeight) / float64(len(weightColors))

	ret := &strings.Builder{}
	for _, s := range p {
		colorI := int((s.Weight - minWeight) / step)
		if colorI < 0 {
			colorI = 0
		}
		if colorI > len(weightColors)-1 {
			colorI = len(weightColors) - 1
		}

		ret.WriteString(weightColors[colorI] + s.Content + resetColor)
		ret.WriteString(s.Suffix)
	}
	return ret.String()
}

// Copy 创建一个拷贝
func (p PromptSentences) Copy() PromptSentences {
	if p == nil {
		return nil
	}
	ret := make(PromptSentences, len(p))
	copy(ret, p)
	return ret
}
