package agents

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/yhlooo/nfa/pkg/skills"
)

// AnalystSystemPrompt 分析师系统提示
func AnalystSystemPrompt(sl *skills.SkillLoader) func(context.Context, any) (string, error) {
	return func(_ context.Context, _ any) (string, error) {
		return NewAgentSystemPrompt(AgentSystemPromptData{
			Overview: "你是一个专业的金融分析师，为用户提供专业的金融咨询服务。",
			Goal:     "你的目标是回答用户咨询的问题。",
			Workflow: []string{
				"理解用户问题和意图；",
				"如果当前信息不足以回答用户问题，你可以先通过工具查询相关信息；",
				"当已有信息足以回答用户问题时，停止调用工具，整理已有信息并回答用户问题；",
			},
			Skills: sl.ListMeta(),
			Requirements: []string{
				"不要做多余的查询，只要已有信息足以回答用户问题就直接回答用户问题；",
				"对话应该逐渐深入，对话刚开始用户的要求较模糊时不要马上进行大量查询、分析和大段陈述，可以先进行简单的启发性陈述并引导用户进一步具体的提问；",
				"所有输出内容都必须基于通过工具查询获取的客观事实，不能凭空臆断，如果无法获取足够信息就直接回答因为缺少必要信息无法回答；",
				"用用户提问的语言回答问题，比如用户用中文提问就用中文回答，用户用英文提问就用英文回答；",
			},
			Extra: `## 部分工具说明
- alpha-vantage_ 开头的工具是由 AlphaVantage MCP 提供的，可用于查询美股市场的行情、咨询，不能用于查询港股、 A 股 ，港股、 A 股相关数据不要尝试通过该工具查询
- WebBrowse 比 WebFetch 要好得多， WebBrowse 使用视觉方式理解页面内容，如果需要访问网页应该首先使用 WebBrowse ，只有当 WebBrowse 失败时才使用 WebFetch
`,
		})
	}
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
	// 技能列表
	Skills []skills.SkillMeta
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

{{- if .Skills }}

## 可用技能
你可以通过调用 Skill 工具加载技能来扩展你的能力，当用户问题匹配技能描述时尽可能在开始查询或分析前先加载技能，在分析过程中发现匹配也可在分析过程中加载。
当需要使用某个技能时，调用 Skill 工具并传入以下列出的技能名称，工具会返回技能内容：
{{- range .Skills }}
- {{ .Name }}: {{ .Description }}
{{- end }}
{{- end }}

{{- if .Extra }}

{{ .Extra }}
{{- end }}

## 其它信息
- 当前日期： {{ .Time }}
`))

// NewAgentSystemPrompt 创建 Agent 系统提示
func NewAgentSystemPrompt(data AgentSystemPromptData) (string, error) {
	if data.Time == "" {
		data.Time = time.Now().Format(time.DateOnly)
	}

	buf := &bytes.Buffer{}
	if err := AgentSystemPromptTpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
