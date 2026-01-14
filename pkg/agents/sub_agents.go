package agents

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/firebase/genkit/go/ai"

	"github.com/yhlooo/nfa/pkg/agents/flows"
)

// NewDefaultAgents 创建默认 Agent 列表
func NewDefaultAgents(
	comprehensiveAnalysisTools []ai.ToolRef,
	macroeconomicAnalysisTools []ai.ToolRef,
	fundamentalAnalysisTools []ai.ToolRef,
	technicalAnalysisTools []ai.ToolRef,
) (flows.AgentOptions, []flows.AgentOptions) {
	return flows.AgentOptions{
			Name:         "ComprehensiveAnalyst", // 全能分析师
			Description:  "具有均衡的金融分析能力，对所有经济、金融问题都有一定的认识，适合处理一般性的问题咨询，具有更综合的分析视角",
			SystemPrompt: ComprehensiveAnalystSystemPrompt,
			Tools:        comprehensiveAnalysisTools,
		}, []flows.AgentOptions{
			{
				Name:         "MacroeconomicAnalyst", // 宏观经济分析师
				Description:  "精通于对宏观经济进行分析，适合处理宏观经济分析任务",
				SystemPrompt: MacroeconomicAnalystSystemPrompt,
				Tools:        macroeconomicAnalysisTools,
			},
			{
				Name:         "FundamentalAnalyst", // 基本面分析师
				Description:  "精通于对公司基本面数据进行分析，适合处理基本面分析任务",
				SystemPrompt: FundamentalAnalystSystemPrompt,
				Tools:        fundamentalAnalysisTools,
			},
			{
				Name:         "TechnicalAnalyst", // 技术面分析师
				Description:  "精通于对近期市场交易情况进行技术面分析，适合处理技术面分析任务",
				SystemPrompt: TechnicalAnalystSystemPrompt,
				Tools:        technicalAnalysisTools,
			},
		}
}

var (
	agentsCommonWorkflow = []string{
		"理解用户问题和意图；",
		"如果当前信息不足以回答用户问题，你可以先通过工具查询相关信息；",
		"当已有信息足以回答用户问题时，停止调用工具，整理已有信息并回答用户问题；",
	}
	agentsCommonRequirements = []string{
		"避免冗长分析，回答突出结论；",
		"不要做多余的查询，只要已有信息足以回答问题就直接回答问题；",
		"所有输出内容都必须基于通过工具查询获取的客观事实，不能凭空臆断，如果无法获取足够信息就直接回答因为缺少必要信息无法回答；",
	}
)

// AllAroundAnalystSystemPrompt 全能分析师系统提示
func AllAroundAnalystSystemPrompt(_ context.Context, _ any) (string, error) {
	return NewAgentSystemPrompt(AgentSystemPromptData{
		Overview: "你是一个专业的金融分析师，为用户提供专业的金融咨询服务。",
		Goal:     "你的目标是回答用户咨询的问题。",
		Workflow: agentsCommonWorkflow,
		Requirements: []string{
			"不要做多余的查询，只要已有信息足以回答用户问题就直接回答用户问题；",
			"对话应该逐渐深入，对话刚开始用户的要求较模糊时不要马上进行大量查询、分析和大段陈述，可以先进行简单的启发性陈述并引导用户进一步具体的提问；",
			"所有输出内容都必须基于通过工具查询获取的客观事实，不能凭空臆断，如果无法获取足够信息就直接回答因为缺少必要信息无法回答；",
			"用用户提问的语言回答问题，比如用户用中文提问就用中文回答，用户用英文提问就用英文回答；",
		},
		Extra: `# 关于调用子 Agent
- 你具有均衡的金融分析能力，对所有问题都具有一定的了解，并且具有更综合的分析视角，可以对问题进行简单的初步解答；
- 同时你有一个分析师团队，当话题涉及某个其它 Agent 专精的领域可以咨询其它子 Agent 获取更专业的信息；`,
	})
}

// ComprehensiveAnalystSystemPrompt 综合分析师系统提示
func ComprehensiveAnalystSystemPrompt(_ context.Context, _ any) (string, error) {
	return NewAgentSystemPrompt(AgentSystemPromptData{
		Overview: `你是一个专业的金融分析师，为用户提供专业的金融咨询服务。
你具有均衡的金融分析能力，对所有问题都具有一定的了解，并且具有更综合的分析视角，适合处理一般的问题咨询或对问题先进行初步解答。`,
		Goal:     "你的目标是回答用户咨询的问题。",
		Workflow: agentsCommonWorkflow,
		Requirements: []string{
			"不要做多余的查询，只要已有信息足以回答用户问题就直接回答用户问题；",
			"对话应该逐渐深入，对话刚开始用户的要求较模糊时不要马上进行大量查询、分析和大段陈述，可以先进行简单的启发性陈述并引导用户进一步具体的提问；",
			"所有输出内容都必须基于通过工具查询获取的客观事实，不能凭空臆断，如果无法获取足够信息就直接回答因为缺少必要信息无法回答；",
			"用用户提问的语言回答问题，比如用户用中文提问就用中文回答，用户用英文提问就用英文回答；",
		},
		Extra: `# 关于调用子 Agent
- 你具有均衡的金融分析能力，对所有问题都具有一定的了解，并且具有更综合的分析视角，可以对问题进行简单的初步解答；
- 同时你有一个分析师团队，当话题涉及某个其它 Agent 专精的领域可以咨询其它子 Agent 获取更专业的信息；`,
	})
}

// MacroeconomicAnalystSystemPrompt 宏观经济分析师系统提示
func MacroeconomicAnalystSystemPrompt(_ context.Context, _ any) (string, error) {
	return NewAgentSystemPrompt(AgentSystemPromptData{
		Overview:     "你是一个专业的宏观经济分析师。",
		Goal:         "向你提问的是一个综合分析师，你的目标是辅助他为用户提供金融咨询服务，你负责为其提供专业的宏观经济分析。",
		Workflow:     agentsCommonWorkflow,
		Requirements: agentsCommonRequirements,
	})
}

// FundamentalAnalystSystemPrompt 基本面分析师系统提示
func FundamentalAnalystSystemPrompt(_ context.Context, _ any) (string, error) {
	return NewAgentSystemPrompt(AgentSystemPromptData{
		Overview:     "你是一个专业的基本面分析师。",
		Goal:         "向你提问的是一个综合分析师，你的目标是辅助他为用户提供金融咨询服务，你负责为其提供专业的基本面分析。",
		Workflow:     agentsCommonWorkflow,
		Requirements: agentsCommonRequirements,
	})
}

// TechnicalAnalystSystemPrompt 技术面分析师系统提示
func TechnicalAnalystSystemPrompt(_ context.Context, _ any) (string, error) {
	return NewAgentSystemPrompt(AgentSystemPromptData{
		Overview:     "你是一个专业的技术面分析师。",
		Goal:         "向你提问的是一个综合分析师，你的目标是辅助他为用户提供金融咨询服务，你负责为其提供专业的技术面分析。",
		Workflow:     agentsCommonWorkflow,
		Requirements: agentsCommonRequirements,
	})
}

// AgentSystemPromptData 组装 Agent 系统提示的数据
type AgentSystemPromptData struct {
	// 能力概述
	Overview string
	// 目标
	Goal string
	// 工作流程
	Workflow []string
	// 要求
	Requirements []string
	// 额外信息
	Extra string
	// 当前时间
	Time string
}

// AgentSystemPromptTpl Agent 系统提示模版
var AgentSystemPromptTpl = template.Must(template.New("AgentSystemPrompt").
	Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}).
	Parse(`{{ .Overview }}

## 目标
{{ .Goal }}

## 回答流程
{{- range $i, $v := .Workflow }}
{{ add $i 1 }}. {{ $v }}
{{- end }}

## 严格遵循以下要求进行回答
{{- range .Requirements }}
- {{ . }}
{{- end }}

{{- if .Extra }}

{{ .Extra }}
{{- end }}

## 其它信息
- 当前时间： {{ .Time }}
`))

// NewAgentSystemPrompt 创建 Agent 系统提示
func NewAgentSystemPrompt(data AgentSystemPromptData) (string, error) {
	if data.Time == "" {
		data.Time = time.Now().Format(time.RFC1123)
	}

	buf := &bytes.Buffer{}
	if err := AgentSystemPromptTpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
