package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSkillTool_Success tests successful skill retrieval
func TestSkillTool_Success(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	skillContent := `---
name: get-price
description: Get asset prices
---

1. Confirm code
2. Query prices
`
	if err := os.Mkdir(filepath.Join(tmpDir, "get-price"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "get-price", SkillFileName), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 加载技能
	if err := loader.Load(t.Context()); err != nil {
		t.Fatal(err)
	}

	// 测试获取技能内容
	result, err := loader.GetSkillContent("get-price")
	if err != nil {
		t.Fatalf("GetSkillContent() error = %v", err)
	}

	if result != skillContent {
		t.Errorf("expected content '%s', got '%s'", skillContent, result)
	}
}

// TestSkillTool_SkillNotFound tests retrieving a non-existent skill
func TestSkillTool_SkillNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 加载技能（空）
	if err := loader.Load(t.Context()); err != nil {
		t.Fatal(err)
	}

	// 测试获取不存在的技能
	_, err := loader.GetSkillContent("non-existent")
	if err == nil {
		t.Fatal("expected error for non-existent skill, got nil")
	}

	expectedErr := "skill 'non-existent' not found"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestSkillTool_MultipleSkills tests retrieving multiple skills
func TestSkillTool_MultipleSkills(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 创建多个技能
	skills := map[string]string{
		"get-price": `---
name: get-price
description: Get asset prices
---

1. Confirm code
2. Query prices
`,
		"analyze-trend": `---
name: analyze-trend
description: Analyze market trends
---

1. Get historical data
2. Calculate indicators
`,
		"send-report": `---
name: send-report
description: Send analysis report
---

1. Format report
2. Send to user
`,
	}

	for name, content := range skills {
		skillDir := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, SkillFileName), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 加载技能
	if err := loader.Load(t.Context()); err != nil {
		t.Fatal(err)
	}

	// 测试每个技能
	for name, expectedContent := range skills {
		result, err := loader.GetSkillContent(name)
		if err != nil {
			t.Fatalf("GetSkillContent() for '%s' error = %v", name, err)
		}

		if result != expectedContent {
			t.Errorf("for skill '%s': expected content '%s', got '%s'", name, expectedContent, result)
		}
	}
}

// TestSkillTool_MetadataError tests skill with invalid metadata
func TestSkillTool_MetadataError(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 创建没有必需字段的技能
	skillDir := filepath.Join(tmpDir, "invalid-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	skillContent := `---
description: Missing name field
---

Content
`
	if err := os.WriteFile(filepath.Join(skillDir, SkillFileName), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 加载技能（应该跳过无效技能）
	if err := loader.Load(t.Context()); err != nil {
		t.Fatal(err)
	}

	// 验证技能没有被加载
	skills := loader.List()
	if len(skills) != 0 {
		t.Errorf("expected 0 skills (invalid metadata), got %d", len(skills))
	}

	// 验证直接获取会失败
	_, err := loader.GetSkillMetadata("invalid-skill")
	if err == nil {
		t.Error("expected error for invalid skill metadata, got nil")
	}
}
