package flows

import (
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
)

// TestTopicClassificationPrompt 测试 TopicClassificationPrompt 方法
func TestTopicClassificationPrompt(t *testing.T) {
	a := assert.New(t)

	ret, err := TopicRoutingPrompt(TopicRoutingInput{
		Messages: []*ai.Message{
			ai.NewUserTextMessage("question 1"),
			ai.NewModelTextMessage("answer 1"),
			ai.NewUserTextMessage("question 2"),
			ai.NewModelTextMessage("answer 2"),
			ai.NewUserTextMessage("question 3"),
			ai.NewModelTextMessage("answer 3"),
			ai.NewUserTextMessage("question 4"),
		},
	})
	a.NoError(err)

	a.Contains(ret, `user:
question 1
model:
answer 1
user:
question 2
model:
answer 2
user:
question 3
model:
answer 3
user:
question 4
`, "Result:\n"+ret)
}
