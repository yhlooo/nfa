package flows

import (
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// AgentOptions Agent 选项
type AgentOptions struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	SystemPrompt ai.PromptFn  `json:"-"`
	Tools        []ai.ToolRef `json:"-"`
}

// DefineMultiAgentsChatFlow 定义多 Agent 对话流程
func DefineMultiAgentsChatFlow(
	g *genkit.Genkit,
	name string,
	mainAgent AgentOptions,
	subAgents []AgentOptions,
) ChatFlow {
	callAgentTool := DefineCallSubAgentTool(g, name+"_CallSubAgent", subAgents)
	return DefineSimpleChatFlow(g, name, FixedGenerateOptions(
		ai.WithSystemFn(mainAgent.SystemPrompt),
		ai.WithTools(append([]ai.ToolRef{callAgentTool}, mainAgent.Tools...)...),
	))
}

// CallSubAgentInput 调用子 Agent 输入
type CallSubAgentInput struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

// CallSubAgentOutput 调用子 Agent 输出
type CallSubAgentOutput struct {
	Messages []*ai.Message `json:"messages"`
	Error    string        `json:"error,omitempty"`
}

// DefineCallSubAgentTool 注册调用子 Agent 工具
func DefineCallSubAgentTool(g *genkit.Genkit, name string, agents []AgentOptions) ai.ToolRef {
	desc := fmt.Sprintf("咨询子 Agent 。可使用的子 Agent ：")
	for _, agent := range agents {
		desc += fmt.Sprintf("\n- %s: %s", agent.Name, agent.Description)
	}

	agentsMap := make(map[string]ChatFlow, len(agents))
	for _, agent := range agents {
		var opts []ai.GenerateOption
		if agent.SystemPrompt != nil {
			opts = append(opts, ai.WithSystemFn(agent.SystemPrompt))
		}
		if agent.Tools != nil {
			opts = append(opts, ai.WithTools(agent.Tools...))
		}
		agentsMap[agent.Name] = DefineSimpleChatFlow(g, name+"_"+agent.Name, FixedGenerateOptions(opts...))
	}

	return genkit.DefineTool(g, name, desc, func(ctx *ai.ToolContext, input CallSubAgentInput) (CallSubAgentOutput, error) {
		agentChatFlow, ok := agentsMap[input.Name]
		if !ok {
			return CallSubAgentOutput{Error: fmt.Sprintf("agent %q not found", input.Name)}, nil
		}
		subAgentChatOut, err := agentChatFlow.Run(
			ctx,
			ChatInput{Prompt: input.Prompt},
		)
		if err != nil {
			return CallSubAgentOutput{Error: err.Error()}, nil
		}
		return CallSubAgentOutput{Messages: subAgentChatOut.Messages}, nil
	})
}
