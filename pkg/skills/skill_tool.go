package skills

import (
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const (
	// LoadSkillToolName 加载技能工具名
	LoadSkillToolName = "Skill"
)

// LoadSkillInput 加载技能输入
type LoadSkillInput struct {
	// 技能名称
	Name string `json:"name"`
}

// LoadSkillOutput 加载技能输出
type LoadSkillOutput struct {
	// 技能名
	Name string `json:"name"`
	// 技能描述
	Description string `json:"description"`
	// 技能内容
	Content string `json:"content,omitempty"`
}

// DefineSkillTool 定义技能工具
func (sl *SkillLoader) DefineSkillTool(g *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(g, LoadSkillToolName,
		`Retrieves the content of a custom skill by name.

Skills are user-defined capabilities stored in ~/.nfa/skills/<skill-name>/SKILL.md.
Each skill contains a YAML frontmatter with name and description, followed by the skill's implementation details.

Input:
- name (string, required): The name of the skill to retrieve

Output:
- name (string): Skill name
- description (string): Description of the skill
- content (string): The SKILL.md content

Example usage:
{
  "name": "get-price"
}`,
		func(ctx *ai.ToolContext, in LoadSkillInput) (LoadSkillOutput, error) {
			// 验证输入
			if in.Name == "" {
				return LoadSkillOutput{}, fmt.Errorf("skill name is required")
			}

			// 获取技能内容
			skill, err := sl.Get(in.Name)
			if err != nil {
				return LoadSkillOutput{}, fmt.Errorf("get skill %q error: %w", in.Name, err)
			}

			return LoadSkillOutput{
				Name:        skill.Meta.Name,
				Description: skill.Meta.Description,
				Content:     skill.Content,
			}, nil
		},
	)
}
