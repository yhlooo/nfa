package dataproviders

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"
)

const (
	AlphaVantageMCPBaseURL = "https://mcp.alphavantage.co/mcp"
)

// AlphaVantageOptions AlphaVantage 选项
type AlphaVantageOptions struct {
	APIKey string `json:"apiKey"`
}

// RegisterTools 注册工具
func (opts *AlphaVantageOptions) RegisterTools(ctx context.Context, g *genkit.Genkit) ([]ai.ToolRef, error) {
	client, err := mcp.NewGenkitMCPClient(mcp.MCPClientOptions{
		Name: "alpha-vantage",
		StreamableHTTP: &mcp.StreamableHTTPConfig{
			BaseURL: AlphaVantageMCPBaseURL,
			Headers: map[string]string{
				"Authorization": "Bearer " + opts.APIKey,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("init alpha vantage mcp client error: %w", err)
	}

	tools, err := client.GetActiveTools(ctx, g)
	if err != nil {
		return nil, fmt.Errorf("get active tools error: %w", err)
	}

	ret := make([]ai.ToolRef, len(tools))
	for i, tool := range tools {
		desc := tool.Definition()
		if len(desc.InputSchema) > 0 {
			genkit.DefineToolWithInputSchema(g, desc.Name, desc.Description, desc.InputSchema, MCPToolFn(tool.RunRaw))
		} else {
			genkit.DefineTool(g, desc.Name, desc.Description, MCPToolFn(tool.RunRaw))
		}
		ret[i] = tool
	}

	return ret, nil
}
