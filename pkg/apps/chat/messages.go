package chat

import "github.com/coder/acp-go-sdk"

type QuitError struct {
	Error error
}

type PromptResult struct {
	Response acp.PromptResponse
	Error    error
}
