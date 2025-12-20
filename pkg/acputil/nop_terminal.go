package acputil

import (
	"context"
	"fmt"

	"github.com/coder/acp-go-sdk"
)

// ClientTerminal ACP 客户端终端
type ClientTerminal interface {
	// CreateTerminal 创建终端
	CreateTerminal(ctx context.Context, params acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error)
	// TerminalOutput 获取终端输出
	TerminalOutput(ctx context.Context, params acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error)
	// ReleaseTerminal 释放终端
	ReleaseTerminal(ctx context.Context, params acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error)
	// WaitForTerminalExit 等待终端结束
	WaitForTerminalExit(ctx context.Context, params acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error)
	// KillTerminalCommand 杀终端
	KillTerminalCommand(ctx context.Context, params acp.KillTerminalCommandRequest) (acp.KillTerminalCommandResponse, error)
}

// NopTerminal 无终端
type NopTerminal struct{}

var _ ClientTerminal = NopTerminal{}

// CreateTerminal 创建终端
func (NopTerminal) CreateTerminal(_ context.Context, _ acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	return acp.CreateTerminalResponse{}, fmt.Errorf("%w: method terminal/create not supported", ErrNotSupported)
}

// TerminalOutput 获取终端输出
func (NopTerminal) TerminalOutput(_ context.Context, _ acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	return acp.TerminalOutputResponse{}, fmt.Errorf("%w: method terminal/output not supported", ErrNotSupported)
}

// ReleaseTerminal 释放终端
func (NopTerminal) ReleaseTerminal(
	_ context.Context,
	_ acp.ReleaseTerminalRequest,
) (acp.ReleaseTerminalResponse, error) {
	return acp.ReleaseTerminalResponse{}, fmt.Errorf("%w: method terminal/release not supported", ErrNotSupported)
}

// WaitForTerminalExit 等待终端结束
func (NopTerminal) WaitForTerminalExit(
	_ context.Context,
	_ acp.WaitForTerminalExitRequest,
) (acp.WaitForTerminalExitResponse, error) {
	return acp.WaitForTerminalExitResponse{},
		fmt.Errorf("%w: method terminal/wait_for_exit not supported", ErrNotSupported)
}

// KillTerminalCommand 杀终端
func (NopTerminal) KillTerminalCommand(
	_ context.Context,
	_ acp.KillTerminalCommandRequest,
) (acp.KillTerminalCommandResponse, error) {
	return acp.KillTerminalCommandResponse{}, fmt.Errorf("%w: method terminal/kill not supported", ErrNotSupported)
}
