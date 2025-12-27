package flows

import (
	"context"
	"fmt"
	"slices"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// SimpleChatFlowName 简单对话流名
const SimpleChatFlowName = "SimpleChat"

// DefineSimpleChatFlow 定义简单对话流
func DefineSimpleChatFlow(g *genkit.Genkit) ChatFlow {
	return genkit.DefineStreamingFlow(g, SimpleChatFlowName,
		func(ctx context.Context, in ChatInput, handleStream core.StreamCallback[*ai.ModelResponseChunk]) (ChatOutput, error) {
			output := ChatOutput{}
			messages := slices.Clone(in.History)
			promptMsg := ai.NewUserTextMessage(in.Prompt)
			messages = append(messages, promptMsg)
			output.Messages = append(output.Messages, promptMsg)

			opts := []ai.GenerateOption{
				ai.WithSystemFn(DefaultChatSystemPrompt),
				ai.WithReturnToolRequests(true),
			}
			if in.ModelName != "" {
				opts = append(opts, ai.WithModelName(in.ModelName))
			}
			if len(in.Tools) > 0 {
				opts = append(opts, ai.WithTools(in.Tools...))
			}
			if handleStream != nil {
				opts = append(opts, ai.WithStreaming(handleTextStream(handleStream)))
			}

			for {
				// 进行一轮生成
				resp, err := genkit.Generate(
					ctx, g,
					append([]ai.GenerateOption{ai.WithMessages(messages...)}, opts...)...,
				)
				if err != nil {
					return output, err
				}
				messages = append(messages, resp.Message)
				output.Messages = append(output.Messages, resp.Message)

				toolRequests := resp.ToolRequests()
				if len(toolRequests) == 0 {
					// 结束一轮对话
					return output, nil
				}

				// 调用工具
				var parts []*ai.Part
				for _, toolReq := range toolRequests {
					if handleStream != nil {
						if err := handleStream(ctx, &ai.ModelResponseChunk{
							Content: []*ai.Part{ai.NewToolRequestPart(toolReq)},
							Role:    resp.Message.Role,
						}); err != nil {
							return output, fmt.Errorf("handle stream error: %w", err)
						}
					}

					toolResp := handleToolCall(ctx, g, toolReq)
					parts = append(parts, toolResp)

					if handleStream != nil {
						if err := handleStream(ctx, &ai.ModelResponseChunk{
							Content: []*ai.Part{toolResp},
							Role:    ai.RoleTool,
						}); err != nil {
							return output, fmt.Errorf("handle stream error: %w", err)
						}
					}
				}
				toolRespMessage := ai.NewMessage(ai.RoleTool, nil, parts...)
				messages = append(messages, toolRespMessage)
				output.Messages = append(output.Messages, toolRespMessage)
			}
		},
	)
}

// handleTextStream 处理文本流
func handleTextStream(handler ai.ModelStreamCallback) ai.ModelStreamCallback {
	return func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
		content := make([]*ai.Part, 0, len(chunk.Content))
		for _, part := range chunk.Content {
			if part.IsReasoning() || part.IsText() || part.IsData() {
				content = append(content, part)
			}
		}
		chunk.Content = content
		if len(chunk.Content) == 0 {
			return nil
		}
		return handler(ctx, chunk)
	}
}

// handleToolCall 处理工具调用
func handleToolCall(ctx context.Context, g *genkit.Genkit, req *ai.ToolRequest) *ai.Part {
	tool := genkit.LookupTool(g, req.Name)
	if tool == nil {
		// 找不到工具
		return ai.NewToolResponsePart(&ai.ToolResponse{
			Name:   req.Name,
			Ref:    req.Ref,
			Output: ToolCallError{Err: fmt.Sprintf("tool %q not found", req.Name)},
		})
	}

	output, err := tool.RunRaw(ctx, req.Input)
	if err != nil {
		return ai.NewToolResponsePart(&ai.ToolResponse{
			Name:   req.Name,
			Ref:    req.Ref,
			Output: ToolCallError{Err: fmt.Sprintf("call tool %q error: %s", req.Name, err.Error())},
		})
	}

	return ai.NewToolResponsePart(&ai.ToolResponse{
		Name:   req.Name,
		Ref:    req.Ref,
		Output: output,
	})
}
