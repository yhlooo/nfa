package fs

import (
	"fmt"
	"io"
	"os"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const (
	// ReadToolName 文件读取工具名
	ReadToolName = "Read"
	// MaxReadSize 最大读取大小 1MB
	MaxReadSize = 1 << 20
)

// ReadInput 文件读取输入
type ReadInput struct {
	// 文件路径（绝对路径或相对路径）
	Path string `json:"path"`
	// 可选：读取的起始字节位置
	Offset int64 `json:"offset,omitempty"`
	// 可选：读取的最大字节数，0 表示使用默认值 1MB，最大不能超过 1MB
	Limit int64 `json:"limit,omitempty"`
}

// ReadOutput 文件读取输出
type ReadOutput struct {
	// 文件内容
	Content string `json:"content"`
	// 文件大小（字节）
	Size int64 `json:"size"`
	// 实际读取的字节数
	BytesRead int64 `json:"bytesRead"`
	// 是否被截断（文件更大但未读完）
	Truncated bool `json:"truncated,omitempty"`
}

// DefineReadTool 定义文件读取工具
func DefineReadTool(g *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(g, ReadToolName, `Read a file from the local filesystem.

以 JSON 格式输入：
- **path**: (string,required) 文件路径，可以是绝对路径或相对路径
- **offset**: (int64,optional) 开始读取的字节位置，默认为 0（文件开头）
- **limit**: (int64,optional) 读取的最大字节数，默认为 1MB，最大允许 1MB。设置为 0 表示使用默认值

输出：
- **content**: 文件内容（文本或二进制）
- **size**: 文件总大小（字节）
- **bytesRead**: 实际读取的字节数
- **truncated**: 文件是否被截断（还有更多内容未读取）

支持的文件类型：不限制文件类型。文本文件可直接阅读，二进制文件将返回原始字节（可能不可读）。

注意：大文件将在读取限制处截断。可使用 offset 和 limit 读取文件的特定部分。
`,
		func(ctx *ai.ToolContext, input ReadInput) (ReadOutput, error) {
			// 2.7 输入验证
			if input.Path == "" {
				return ReadOutput{}, fmt.Errorf("path is required")
			}
			if input.Offset < 0 {
				return ReadOutput{}, fmt.Errorf("offset must be >= 0, got %d", input.Offset)
			}
			if input.Limit < 0 {
				return ReadOutput{}, fmt.Errorf("limit must be >= 0, got %d", input.Limit)
			}
			if input.Limit > MaxReadSize {
				return ReadOutput{}, fmt.Errorf("limit must be <= %d, got %d", MaxReadSize, input.Limit)
			}

			// 2.8 打开文件
			file, err := os.Open(input.Path)
			if err != nil {
				return ReadOutput{}, fmt.Errorf("open file %q error: %w", input.Path, err)
			}
			defer func() { _ = file.Close() }()

			// 2.9 获取文件信息
			fileInfo, err := file.Stat()
			if err != nil {
				return ReadOutput{}, fmt.Errorf("stat file %q error: %w", input.Path, err)
			}
			fileSize := fileInfo.Size()

			// 2.10 定位到 offset
			if input.Offset > 0 {
				_, err = file.Seek(input.Offset, io.SeekStart)
				if err != nil {
					return ReadOutput{}, fmt.Errorf("seek to offset %d error: %w", input.Offset, err)
				}
			}

			// 2.11 限制读取
			effectiveLimit := input.Limit
			if effectiveLimit == 0 {
				effectiveLimit = MaxReadSize
			}
			limitedReader := io.LimitReader(file, effectiveLimit)

			content, err := io.ReadAll(limitedReader)
			if err != nil {
				return ReadOutput{}, fmt.Errorf("read file %q error: %w", input.Path, err)
			}
			bytesRead := int64(len(content))

			// 2.12 检测截断
			truncated := false
			if bytesRead == effectiveLimit {
				// 尝试多读 1 字节确认是否还有内容
				oneByte := make([]byte, 1)
				n, _ := file.Read(oneByte)
				if n > 0 {
					truncated = true
				}
			}

			return ReadOutput{
				Content:   string(content),
				Size:      fileSize,
				BytesRead: bytesRead,
				Truncated: truncated,
			}, nil
		},
	)
}
