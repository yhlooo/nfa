package skills

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"gopkg.in/yaml.v3"
)

// SkillMetadata 技能元数据
type SkillMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// SkillContent 技能内容
type SkillContent struct {
	Metadata SkillMetadata
	Content  string
}

// ParseSkillFile 解析技能文件
func ParseSkillFile(path string) (*SkillContent, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.EACCES) {
			return nil, fmt.Errorf("permission denied reading skill file '%s': %w", path, err)
		}
		return nil, fmt.Errorf("read skill file error: %w", err)
	}

	return ParseSkillContent(string(content))
}

// ParseSkillContent 解析技能内容
func ParseSkillContent(content string) (*SkillContent, error) {
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
	var metadata SkillMetadata
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return nil, fmt.Errorf("parse YAML frontmatter error: %w", err)
	}

	// 验证必填字段
	if metadata.Name == "" {
		return nil, fmt.Errorf("missing required field: name")
	}
	if metadata.Description == "" {
		return nil, fmt.Errorf("missing required field: description")
	}

	// 清理 markdown 内容（去除开头的换行）
	markdownContent = strings.TrimLeft(markdownContent, "\n\r")

	return &SkillContent{
		Metadata: metadata,
		Content:  markdownContent,
	}, nil
}

// GetSkillMetadata 获取技能元数据
func (sl *SkillLoader) GetSkillMetadata(name string) (*SkillMetadata, error) {
	skill, ok := sl.Get(name)
	if !ok {
		return nil, fmt.Errorf("skill '%s' not found", name)
	}

	content, err := ParseSkillFile(skill.Path)
	if err != nil {
		return nil, fmt.Errorf("parse skill '%s' error: %w", name, err)
	}

	return &content.Metadata, nil
}

// GetSkillContent 获取技能完整内容（包括 frontmatter）
func (sl *SkillLoader) GetSkillContent(name string) (string, error) {
	skill, ok := sl.Get(name)
	if !ok {
		return "", fmt.Errorf("skill '%s' not found", name)
	}

	content, err := os.ReadFile(skill.Path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.EACCES) {
			return "", fmt.Errorf("permission denied reading skill '%s': %w", name, err)
		}
		return "", fmt.Errorf("read skill file error: %w", err)
	}

	return string(content), nil
}

// FormatSkillContent 格式化技能内容用于显示
func FormatSkillContent(content *SkillContent) string {
	var buf bytes.Buffer

	// 写入 frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("name: %s\n", content.Metadata.Name))
	buf.WriteString(fmt.Sprintf("description: %s\n", content.Metadata.Description))
	buf.WriteString("---\n\n")

	// 写入内容
	buf.WriteString(content.Content)
	// 如果内容不以换行符结尾，添加一个
	if content.Content != "" && !strings.HasSuffix(content.Content, "\n") {
		buf.WriteString("\n")
	}

	return buf.String()
}
