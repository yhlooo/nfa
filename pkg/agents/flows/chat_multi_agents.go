package flows

import (
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type AgentOptions struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	SystemPrompt ai.PromptFn  `json:"-"`
	Tools        []ai.ToolRef `json:"-"`
}

// DefineMultiAgentsChatFlow 定义多 Agent 对话流程
//
// NOTE: 该 flow 只能被单线程调用，否则 agent 切换会产生冲突
func DefineMultiAgentsChatFlow(g *genkit.Genkit, name string, agents []AgentOptions, defaultAgent string) ChatFlow {
	_, genOpts := DefineSwitchAgentTool(g, name+"_SwitchAgent", agents, defaultAgent)
	return DefineSimpleChatFlow(g, name, genOpts)
}

// SwitchAgentInput 切换 Agent 输入
type SwitchAgentInput struct {
	Name string `json:"name"`
}

// SwitchAgentOutput 切换 Agent 输出
type SwitchAgentOutput struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// DefineSwitchAgentTool 定义切换 Agent 工具
func DefineSwitchAgentTool(
	g *genkit.Genkit,
	name string,
	agents []AgentOptions,
	defaultAgent string,
) (ai.Tool, GenerateOptionsFn) {
	currentAgent := defaultAgent
	agentsMap := make(map[string]AgentOptions, len(agents))
	for _, agent := range agents {
		agentsMap[agent.Name] = agent
	}

	desc := fmt.Sprintf("切换接下来对话使用的 Agent 。可切换的 Agent ：")
	for _, agent := range agents {
		desc += fmt.Sprintf("\n- %s: %s", agent.Name, agent.Description)
	}

	tool := genkit.DefineTool(g, name, desc,
		func(ctx *ai.ToolContext, input SwitchAgentInput) (SwitchAgentOutput, error) {
			_, ok := agentsMap[input.Name]
			if !ok {
				return SwitchAgentOutput{OK: false, Error: fmt.Sprintf("agent %q not found", input.Name)}, nil
			}

			currentAgent = input.Name
			return SwitchAgentOutput{OK: true}, nil
		},
	)

	return tool, func() []ai.GenerateOption {
		agentOpts := agentsMap[currentAgent]

		genOpts := []ai.GenerateOption{
			ai.WithTools(append([]ai.ToolRef{tool}, agentOpts.Tools...)...),
		}
		if agentOpts.SystemPrompt != nil {
			genOpts = append(genOpts, ai.WithSystemFn(agentOpts.SystemPrompt))
		}

		return genOpts
	}
}
