package agents

import (
	"context"

	"github.com/firebase/genkit/go/genkit"
)

type MainInput struct {
	Message string `json:"message"`
}

type MainOutput struct {
	Message string `json:"message"`
}

func (a *NFAAgent) registerMainFlow() {
	a.mainFlow = genkit.DefineFlow(a.g, "mainFlow", func(ctx context.Context, in MainInput) (MainOutput, error) {
		return MainOutput{}, nil
	})
}
