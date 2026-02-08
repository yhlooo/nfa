package dualphase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPromptSentences_WithWeightColors 测试 PromptSentences.WithWeightColors 方法
func TestPromptSentences_WithWeightColors(t *testing.T) {
	a := assert.New(t)

	expectedRet := "\x1b[48;5;17m###Task type###\x1b[0m\n" +
		"\x1b[48;5;17mTask type: This is a basic arithmetic calculation task.\x1b[0m\n\n" +
		"\x1b[48;5;17m###Task detailed description###\x1b[0m\n" +
		"\x1b[48;5;17mTask detailed description: Given a simple arithmetic expression involving two " +
		"numbers and one of the four basic operators (+, -, *, /), calculate the result.\x1b[0m\n\n" +
		"\x1b[48;5;17m###Your output must satisfy the following format and constraints###\x1b[0m\n" +
		"\x1b[48;5;18mOutput format(type): A single integer or decimal number.\x1b[0m\n" +
		"\x1b[48;5;17mOutput constraints: The output must be the exact numerical result of the calculation.\x1b[0m " +
		"\x1b[48;5;17mDo not include any additional text, symbols, or explanations.\x1b[0m\n\n" +
		"\x1b[48;5;17m###You must follow the reasoning process###\x1b[0m\n\x1b[48;5;17m1. Identify the two " +
		"numbers and the operator in the input expression.\x1b[0m\n" +
		"\x1b[48;5;21m2. Perform the corresponding arithmetic operation (addition, subtraction, multiplication, " +
		"or division) on the two numbers.\x1b[0m\n\x1b[48;5;17m3. Output the final result as a number.\x1b[0m\n\n" +
		"\x1b[48;5;17m###Tips###\x1b[0m\n\x1b[48;5;17m- Ensure you handle integer and decimal results " +
		"correctly.\x1b[0m\n\x1b[48;5;17m- For division, provide the exact quotient (e.g., 4/2 outputs 2, " +
		"5/2 outputs 2.5).\x1b[0m\n\x1b[48;5;17m- The input will always be a valid, simple expression with " +
		"two operands and one operator.\x1b[0m\n"
	ret := PromptSentences{
		{Sentence: Sentence{Content: "###Task type###", Suffix: "\n"}, Weight: 0},
		{Sentence: Sentence{Content: "Task type: This is a basic arithmetic calculation task.", Suffix: "\n\n"}, Weight: 1},
		{Sentence: Sentence{Content: "###Task detailed description###", Suffix: "\n"}, Weight: 3},
		{Sentence: Sentence{Content: "Task detailed description: Given a simple arithmetic expression involving two numbers and one of the four basic operators (+, -, *, /), calculate the result.", Suffix: "\n\n"}, Weight: 5},
		{Sentence: Sentence{Content: "###Your output must satisfy the following format and constraints###", Suffix: "\n"}, Weight: 7},
		{Sentence: Sentence{Content: "Output format(type): A single integer or decimal number.", Suffix: "\n"}, Weight: 20},
		{Sentence: Sentence{Content: "Output constraints: The output must be the exact numerical result of the calculation.", Suffix: " "}, Weight: 13},
		{Sentence: Sentence{Content: "Do not include any additional text, symbols, or explanations.", Suffix: "\n\n"}, Weight: 6},
		{Sentence: Sentence{Content: "###You must follow the reasoning process###", Suffix: "\n"}, Weight: 7},
		{Sentence: Sentence{Content: "1. Identify the two numbers and the operator in the input expression.", Suffix: "\n"}, Weight: 15},
		{Sentence: Sentence{Content: "2. Perform the corresponding arithmetic operation (addition, subtraction, multiplication, or division) on the two numbers.", Suffix: "\n"}, Weight: 21},
		{Sentence: Sentence{Content: "3. Output the final result as a number.", Suffix: "\n\n"}, Weight: 17},
		{Sentence: Sentence{Content: "###Tips###", Suffix: "\n"}, Weight: 10},
		{Sentence: Sentence{Content: "- Ensure you handle integer and decimal results correctly.", Suffix: "\n"}, Weight: 6},
		{Sentence: Sentence{Content: "- For division, provide the exact quotient (e.g., 4/2 outputs 2, 5/2 outputs 2.5).", Suffix: "\n"}, Weight: 18},
		{Sentence: Sentence{Content: "- The input will always be a valid, simple expression with two operands and one operator.", Suffix: "\n"}, Weight: 16},
	}.WithWeightColors()
	a.Equal(expectedRet, ret)
}
