package flows

import (
	"context"
	"fmt"
	"slices"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/ctxutil"
	"github.com/yhlooo/nfa/pkg/tokentracker"
)

// DefineSimpleChatFlow 定义简单对话流程
func DefineSimpleChatFlow(g *genkit.Genkit, name string, genOpts ...ai.GenerateOption) ChatFlow {
	return genkit.DefineFlow(g, name,
		func(ctx context.Context, in ChatInput) (ChatOutput, error) {
			output := ChatOutput{}
			messages := slices.Clone(in.History)
			promptMsg := ai.NewUserTextMessage(in.Prompt)
			messages = append(messages, promptMsg)

			modelName := ""
			if m, ok := ctxutil.ModelsFromContext(ctx); ok {
				modelName = m.GetPrimary()
			}

			opts := []ai.GenerateOption{
				ai.WithReturnToolRequests(true),
				ai.WithMiddleware(tokentracker.ModelMiddlewareFromContext(ctx, modelName)),
			}
			if modelName != "" {
				opts = append(opts, ai.WithModelName(modelName))
			}
			handleStream := ctxutil.HandleStreamFnFromContext(ctx)
			if handleStream != nil {
				ctx = ctxutil.ContextWithHandleStreamFn(ctx, handleStream)
				opts = append(opts, ai.WithStreaming(handleTextStream(handleStream, true, true)))
			}

			reflected := 0

			for {
				curTurnOpts := append([]ai.GenerateOption{ai.WithMessages(messages...)}, opts...)
				curTurnOpts = append(curTurnOpts, genOpts...)

				// 检查上下文限制
				if output.LastContextWindow > in.MaxContextWindow {
					return output, nil
				}

				// 进行一轮生成
				resp, err := genkit.Generate(ctx, g, curTurnOpts...)
				if err != nil {
					return output, err
				}
				messages = append(messages, resp.Message)
				if resp.Usage != nil {
					output.LastContextWindow = int64(resp.Usage.InputTokens)
				}

				toolRequests := resp.ToolRequests()
				if len(toolRequests) == 0 {
					// 反思一轮
					if reflected < 1 {
						messages = append(messages, ai.NewUserTextMessage(`[SystemPrompt] 请根据以下检查项反思你的回答是否正确解决了用户的问题，并在确认无误后重新组织回答：
1. 形式：检查回答在形式上是否真正回答了用户的问题？
2. 广度和深度：回顾自己的之前的思考是否已经充分考虑了问题的广度和深度，是否有关键遗漏？
3. 事实和数据：评估支撑自己结论的关键数据是什么？数据对结论是否形成有力支撑？这些数据的来源是否真实可靠？
4. 严谨性：回答是否向用户明确澄清回答的局限性、适用范围，以确保不会导致用户产生重大误解？
如果回答存在缺陷请调整或继续思考、探索，如果确认无误则重新组织回答。
**注意：新组织的回答应当作给用户的第一个回答，不应该向用户透露反思结果等额外信息**
`))

						reflected++
						continue
					}

					// 结束对话
					output.Messages = append(output.Messages, resp.Message)
					return output, nil
				}

				output.Messages = append(output.Messages, resp.Message)

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

// pruneReasoning 去除消息中的思考过程
func pruneReasoning(msg *ai.Message) *ai.Message {
	parts := make([]*ai.Part, 0, len(msg.Content))
	for _, part := range msg.Content {
		if part.IsReasoning() {
			continue
		}
		parts = append(parts, part)
	}

	return &ai.Message{
		Content:  parts,
		Metadata: msg.Metadata,
		Role:     msg.Role,
	}
}
