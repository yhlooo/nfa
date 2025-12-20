package tty

import "github.com/coder/acp-go-sdk"

type QuitError error

type PromptResult struct {
	Response acp.PromptResponse
	Error    error
}
