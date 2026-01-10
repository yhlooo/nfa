package dualphase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitializationPrompt 测试 InitializationPrompt 方法
func TestInitializationPrompt(t *testing.T) {
	a := assert.New(t)

	ret, err := InitializationPrompt(InitializationInput{
		TrainingData: []InputOutputPair{
			{Input: "input1", Output: "output1"},
		},
	})
	a.NoError(err)
	a.Contains(ret, `Here are some correct input-output pairs which strictly meet all your requirements:

Input: input1
Output: output1

The instruction given contains the following parts.`)

	ret, err = InitializationPrompt(InitializationInput{})
	a.NoError(err)
	a.Contains(ret, `Here are some correct input-output pairs which strictly meet all your requirements:

The instruction given contains the following parts.`)

	ret, err = InitializationPrompt(InitializationInput{
		TrainingData: []InputOutputPair{
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
