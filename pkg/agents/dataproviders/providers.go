package dataproviders

import (
	"context"

	"github.com/firebase/genkit/go/ai"
)

// DataProvider 数据供应商配置
type DataProvider struct {
	AlphaVantage *AlphaVantageOptions `json:"alphaVantage,omitempty"`
}

// MCPToolFn MCP 工具方法
func MCPToolFn[In any, Out any](fn func(ctx context.Context, input In) (Out, error)) ai.ToolFunc[In, Out] {
	return func(ctx *ai.ToolContext, input In) (Out, error) {
		return fn(ctx, input)
	}
}
