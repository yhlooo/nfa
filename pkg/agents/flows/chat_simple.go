package flows

import (
	"context"
	"fmt"
	"slices"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// DefineSimpleChatFlow 定义简单对话流程
func DefineSimpleChatFlow(g *genkit.Genkit, name string, genOpts GenerateOptionsFn) ChatFlow {
	return genkit.DefineStreamingFlow(g, name,
		func(ctx context.Context, in ChatInput, handleStream ai.ModelStreamCallback) (ChatOutput, error) {
			output := ChatOutput{}
			messages := slices.Clone(in.History)
			promptMsg := ai.NewUserTextMessage(in.Prompt)
			messages = append(messages, promptMsg)

			opts := []ai.GenerateOption{
				ai.WithReturnToolRequests(true),
			}
			if m, ok := ctxutil.ModelsFromContext(ctx); ok {
				opts = append(opts, ai.WithModelName(m.GetMain()))
			}
			if handleStream != nil {
				ctx = ctxutil.ContextWithHandleStreamFn(ctx, handleStream)
				opts = append(opts, ai.WithStreaming(handleTextStream(handleStream, true, true)))
			}

			for {
				curTurnOpts := append([]ai.GenerateOption{ai.WithMessages(messages...)}, opts...)
				curTurnOpts = append(curTurnOpts, genOpts()...)

				// 进行一轮生成
				resp, err := genkit.Generate(ctx, g, curTurnOpts...)
				if err != nil {
					return output, err
				}
				messages = append(messages, resp.Message)
				ctxutil.AddModelUsageToContext(ctx, resp.Usage)

				toolRequests := resp.ToolRequests()
				if len(toolRequests) == 0 {
					// 结束一轮对话
					output.Messages = append(output.Messages, ai.NewModelTextMessage(resp.Text()))
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
			}
		},
	)
}

// handleTextStream 处理文本流
func handleTextStream(handler ai.ModelStreamCallback, reasoning, text bool) ai.ModelStreamCallback {
	return func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
		content := make([]*ai.Part, 0, len(chunk.Content))
		for _, part := range chunk.Content {
			if reasoning && part.IsReasoning() {
				content = append(content, part)
			} else if text && (part.IsText() || part.IsData()) {
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
