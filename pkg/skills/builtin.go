package skills

import (
	"fmt"
	"path"
	"strings"
)

// readBuiltinSkill 从 embed.FS 读取内置技能
func readBuiltinSkill(virtualPath string) (*Skill, error) {
	// virtualPath 格式: "builtin/skill-name"
	// 提取技能名称
	parts := strings.Split(virtualPath, "/")
	if len(parts) != 2 || parts[0] != "builtin" {
		return nil, fmt.Errorf("invalid builtin skill path: %s", virtualPath)
	}
	skillName := parts[1]

	// 从 embed.FS 读取
	content, err := builtinSkillsFS.ReadFile(path.Join(virtualPath, SkillFileName))
	if err != nil {
		return nil, fmt.Errorf("read builtin skill %q error: %w", skillName, err)
	}

	// 解析内容
	skill, err := ParseSkillContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse builtin skill %q error: %w", skillName, err)
	}

	// 设置来源
	skill.Meta.Source = SkillSourceBuiltin

	return skill, nil
}
