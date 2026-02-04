package agents

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewAgentSystemPrompt 测试 NewAgentSystemPrompt 方法
func TestNewAgentSystemPrompt(t *testing.T) {
	a := assert.New(t)

	data := AgentSystemPromptData{
		Overview: "你是一个 XXX",
		Goal:     "目标！目标！",
		Workflow: []string{
			"首先",
			"然后",
			"最后",
		},
		Requirements: []string{
			"要这样、这样",
			"不是那样",
			"是这样",
		},
	}

	ret, err := NewAgentSystemPrompt(data)
	a.NoError(err)
	a.Contains(ret, `你是一个 XXX

## 目标
目标！目标！

## 回答流程
1. 首先
2. 然后
3. 最后

## 严格遵循以下要求进行回答
- 要这样、这样
- 不是那样
- 是这样

## 其它信息
- 当前时间：`, "Result:\n"+ret)

	data.Extra = "额外信息"
	ret, err = NewAgentSystemPrompt(data)
	a.NoError(err)
	a.Contains(ret, `## 严格遵循以下要求进行回答
- 要这样、这样
- 不是那样
- 是这样

额外信息

## 其它信息`, "Result:\n"+ret)
}
