package dualphase

import (
	"math"
	"math/rand/v2"
	"strings"
)

// WeightedSentence 带权重的句子
type WeightedSentence struct {
	Sentence

	// 权重
	Weight float64 `json:"weight,omitempty,string"`
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
		if s.Ignore {
			continue
		}

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
		if s.Ignore {
			ret.WriteString(s.Content + s.Suffix)
			continue
		}

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

// Sample 采样一个句子
func (p PromptSentences) Sample() (int, WeightedSentence) {
	if len(p) == 0 {
		return 0, WeightedSentence{}
	}

	// 根据每个句子权重计算被采样概率
	pr := p.SampleProbability()

	// 根据概率随机选择一个句子
	cursor := rand.Float64()
	sumP := float64(0)
	for i, s := range p {
		sumP += pr[i]
		if sumP > cursor {
			return i, s
		}
	}
	// 因浮点误差有极小概率概率和不为 1 ，兜底选择最后一个句子
	return len(p) - 1, p[len(p)-1]
}

// UpdateWeight 更新权重
//
// 参考文章中 Eq. 9
func (p PromptSentences) UpdateWeight(mixedEvalResult, learningRate float64) {
	pr := p.SampleProbability()

	for i, s := range p {
		p[i].Weight = s.Weight * math.Exp((learningRate*mixedEvalResult)/(pr[i]*float64(len(p))))
	}
}

// SampleProbability 根据每个句子权重计算当前被采样概率
//
// 参考文章中 Eq. 5
func (p PromptSentences) SampleProbability() []float64 {
	pr := make([]float64, len(p))
	expSum := float64(0)
	for _, s := range p {
		if s.Ignore {
			continue
		}
		expSum += math.Exp(s.Weight)
	}
	for i, s := range p {
		if s.Ignore {
			pr[i] = 0
			continue
		}
		pr[i] = math.Exp(s.Weight) / expSum
	}

	return pr
}
