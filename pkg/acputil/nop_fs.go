package acputil

import (
	"context"
	"fmt"

	"github.com/coder/acp-go-sdk"
)

// ClientFS ACP 客户端文件系统
type ClientFS interface {
	// ReadTextFile 读文本文件
	ReadTextFile(ctx context.Context, params acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error)
	// WriteTextFile 写文本文件
	WriteTextFile(ctx context.Context, params acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error)
}

// NopFS 无文件系统
type NopFS struct{}

var _ ClientFS = NopFS{}

// ReadTextFile 读文本文件
func (NopFS) ReadTextFile(_ context.Context, _ acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	return acp.ReadTextFileResponse{}, fmt.Errorf("%w: method fs/read_text_file not supported", ErrNotSupported)
}

// WriteTextFile 写文本文件
func (NopFS) WriteTextFile(_ context.Context, _ acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	return acp.WriteTextFileResponse{}, fmt.Errorf("%w: method fs/write_text_file not supported", ErrNotSupported)
}
