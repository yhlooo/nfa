package skills

import (
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"
)

const (
	// SkillToolName 技能工具名
	SkillToolName = "Skill"
)

// SkillInput 技能工具输入
type SkillInput struct {
	// 技能名称
	Name string `json:"name"`
}

// SkillOutput 技能工具输出
type SkillOutput struct {
	// 技能内容
	Content string `json:"content,omitempty"`
	// 错误信息
	Error string `json:"error,omitempty"`
}

// DefineSkillTool 定义技能工具
func DefineSkillTool(g *genkit.Genkit, loader *SkillLoader) ai.ToolRef {
	return genkit.DefineTool(g, SkillToolName,
		`Retrieves the content of a custom skill by name.

Skills are user-defined capabilities stored in ~/.nfa/skills/<skill-name>/SKILL.md.
Each skill contains a YAML frontmatter with name and description, followed by the skill's implementation details.

Input:
- name (string, required): The name of the skill to retrieve

Output:
- content (string): The full SKILL.md content including frontmatter and markdown
- error (string): Error message if the skill is not found or cannot be read

Example usage:
{
  "name": "get-price"
}`,
		func(ctx *ai.ToolContext, input SkillInput) (SkillOutput, error) {
			logger := logr.FromContextOrDiscard(ctx)

			// 验证输入
			if input.Name == "" {
				logger.Info("Skill tool called without name parameter")
				return SkillOutput{
					Error: "skill name is required",
				}, nil
			}

			// 获取技能内容
			content, err := loader.GetSkillContent(input.Name)
			if err != nil {
				logger.Info(fmt.Sprintf("Skill tool failed to get skill '%s': %v", input.Name, err))
				return SkillOutput{
					Error: fmt.Sprintf("Skill '%s' not found", input.Name),
				}, nil
			}

			logger.Info(fmt.Sprintf("Skill tool successfully retrieved skill '%s'", input.Name))
			return SkillOutput{
				Content: content,
			}, nil
		},
	)
}
