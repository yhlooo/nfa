package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// SkillFileName skill 文件名
	SkillFileName = "SKILL.md"
)

// Skill 技能内容
type Skill struct { // TODO
	Meta    SkillMeta
	Content string
}

// SkillMeta 技能元数据
type SkillMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// ReadSkill 读取技能
func ReadSkill(path string) (*Skill, error) {
	content, err := os.ReadFile(filepath.Join(path, SkillFileName))
	if err != nil {
		return nil, fmt.Errorf("read skill file %q error: %w", SkillFileName, err)
	}
	return ParseSkillContent(string(content))
}

// ParseSkillContent 解析技能内容
func ParseSkillContent(content string) (*Skill, error) {
	// 查找 YAML frontmatter 开始标记
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return nil, fmt.Errorf("skill file must start with YAML frontmatter (---)")
	}

	// 查找 frontmatter 结束标记
	// 使用 strings.Index 从第4个字符开始搜索，找到后相对于原始字符串的位置需要加4
	afterStartMarker := content[4:]
	endIndex := strings.Index(afterStartMarker, "---\n")
	separatorLen := 4
	if endIndex == -1 {
		endIndex = strings.Index(afterStartMarker, "---\r\n")
		separatorLen = 5
	}
	if endIndex == -1 {
		return nil, fmt.Errorf("frontmatter not closed (---)")
	}

	// endIndex 是相对于 afterStartMarker 的位置，所以原始位置是 4+endIndex
	// 提取 frontmatter 内容
	frontmatter := afterStartMarker[:endIndex]
	markdownContent := afterStartMarker[endIndex+separatorLen:] // 跳过结束标记

	// 解析 YAML
	var meta SkillMeta
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return nil, fmt.Errorf("parse YAML frontmatter error: %w", err)
	}

	// 验证必填字段
	if meta.Name == "" {
		return nil, fmt.Errorf("missing required field: name")
	}
	if meta.Description == "" {
		return nil, fmt.Errorf("missing required field: description")
	}

	// 清理 markdown 内容（去除开头的换行）
	markdownContent = strings.TrimLeft(markdownContent, "\n\r")

	return &Skill{
		Meta:    meta,
		Content: markdownContent,
	}, nil
}
