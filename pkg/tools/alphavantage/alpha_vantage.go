package alphavantage

import (
	"context"
	"fmt"
	"net/url"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"

	"github.com/yhlooo/nfa/pkg/tools"
)

const (
	BaseURL = "https://mcp.alphavantage.co/mcp"
)

// Options AlphaVantage 选项
type Options struct {
	APIKey string `json:"apiKey"`
}

// RegisterTools 注册工具
func (opts *Options) RegisterTools(ctx context.Context, g *genkit.Genkit) ([]ai.ToolRef, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf(".apiKey is required")
	}

	client, err := mcp.NewGenkitMCPClient(mcp.MCPClientOptions{
		Name: "alpha-vantage",
		StreamableHTTP: &mcp.StreamableHTTPConfig{
			BaseURL: BaseURL + "?apikey=" + url.QueryEscape(opts.APIKey),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("init alpha vantage mcp client error: %w", err)
	}

	toolList, err := client.GetActiveTools(ctx, g)
	if err != nil {
		return nil, fmt.Errorf("get active tools error: %w", err)
	}

	var allTools []ai.ToolRef
	for _, tool := range toolList {
		desc := tool.Definition()
		var toolOpts []ai.ToolOption
		if len(desc.InputSchema) > 0 {
			toolOpts = append(toolOpts, ai.WithInputSchema(desc.InputSchema))
		}
		genkit.DefineTool(g, desc.Name, desc.Description, tools.MCPToolFn(tool.RunRaw), toolOpts...)

		allTools = append(allTools, tool)
	}

	return allTools, nil
}
