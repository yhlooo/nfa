package tools

import (
	"fmt"
	"io"
	"net/http"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
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
	return genkit.DefineTool(g, "WebFetch", "Retrieve content from the specified URL",
		func(ctx *ai.ToolContext, input WebFetchInput) (WebFetchOutput, error) {
			resp, err := http.Get(input.URL)
			if err != nil {
				return WebFetchOutput{}, fmt.Errorf("send http request error: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			ret := WebFetchOutput{StatusCode: resp.StatusCode}

			content, err := io.ReadAll(io.LimitReader(resp.Body, 100<<20))
			if err != nil {
				return ret, fmt.Errorf("read response body error: %w", err)
			}
			ret.Content = string(content)

			return ret, nil
		},
	)
}
