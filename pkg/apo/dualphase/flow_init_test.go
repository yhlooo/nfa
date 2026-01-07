package dualphase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOriginalInitializationPrompt 测试 OriginalInitializationPrompt 方法
func TestOriginalInitializationPrompt(t *testing.T) {
	a := assert.New(t)

	ret, err := OriginalInitializationPrompt(OriginalInitializationInput{
		TrainingData: []TrainingData{
			{Input: "input1", Output: "output1"},
		},
	})
	a.NoError(err)
	a.Contains(ret, `Here are some correct input-output pairs which strictly meet all your requirements:

Input: input1
Output: output1

The instruction given contains the following parts.`)

	ret, err = OriginalInitializationPrompt(OriginalInitializationInput{})
	a.NoError(err)
	a.Contains(ret, `Here are some correct input-output pairs which strictly meet all your requirements:

The instruction given contains the following parts.`)

	ret, err = OriginalInitializationPrompt(OriginalInitializationInput{
		TrainingData: []TrainingData{
			{Input: "input1", Output: "output1"},
			{Input: "input2", Output: "output2"},
		},
	})
	a.NoError(err)
	a.Contains(ret, `Here are some correct input-output pairs which strictly meet all your requirements:

Input: input1
Output: output1

Input: input2
Output: output2

The instruction given contains the following parts.`)
}
