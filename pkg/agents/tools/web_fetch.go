package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"
)

// WebFetchInput 获取 URL 内容输入
type WebFetchInput struct {
	URL string `json:"url"`
}

// WebFetchOutput 获取 URL 内容输出
type WebFetchOutput struct {
	StatusCode int    `json:"statusCode"`
	Content    string `json:"content"`
}

// DefineWebFetchTool 定义获取 URL 内容工具
func DefineWebFetchTool(g *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(g, "WebFetch",
		`Fetch the specified URL, read its content, convert it into readable text, and output.

The following types of URLs are supported:
- HTTP (http:// or https://)
- File (file://)

The following types of content are supported:
- Plain Text
- HTML
- JSON
- PDF
`,
		func(ctx *ai.ToolContext, input WebFetchInput) (WebFetchOutput, error) {
			logger := logr.FromContextOrDiscard(ctx)

			resp, err := http.Get(input.URL)
			if err != nil {
				return WebFetchOutput{}, fmt.Errorf("send http request error: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			ret := WebFetchOutput{StatusCode: resp.StatusCode}

			var content []byte
			switch resp.Header.Get("Content-Type") {
			case "application/pdf":
				// PDF
				logger.Info("pdf content found")
				content, err = convertPDFToText(ctx, resp.Body)

			default:
				// 其它
				logger.Info(fmt.Sprintf("content type: %s", resp.Header.Get("Content-Type")))
				content, err = io.ReadAll(io.LimitReader(resp.Body, 100<<20))
			}
			if err != nil {
				return ret, fmt.Errorf("read response body error: %w", err)
			}

			ret.Content = string(content)

			return ret, nil
		},
	)
}

// convertPDFToText 将 PDF 转换为文本
func convertPDFToText(ctx context.Context, r io.Reader) ([]byte, error) {
	tmpfile, err := os.CreateTemp("", "nfa-pdf-")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	_, err = io.Copy(tmpfile, r)
	_ = tmpfile.Close()
	if err != nil {
		return nil, fmt.Errorf("read pdf error: %w", err)
	}

	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", tmpfile.Name(), "-")
	out := &bytes.Buffer{}
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("convert pdf to text error: %w", err)
	}

	_ = os.Remove(tmpfile.Name())

	return out.Bytes(), nil
}
