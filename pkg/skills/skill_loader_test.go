package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSkillLoader(t *testing.T) {
	loader := NewSkillLoader("/home/user/skills")
	if loader == nil {
		t.Fatal("NewSkillLoader returned nil")
	}
	expectedPath := "/home/user/skills"
	if loader.skillsDir != expectedPath {
		t.Errorf("expected skills dir %s, got %s", expectedPath, loader.skillsDir)
	}
}

func TestLoad_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	err := loader.Load(t.Context())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	skills := loader.List()
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

func TestLoad_ValidSkill(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 创建技能目录
	skillDir := filepath.Join(loader.skillsDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建 SKILL.md 文件
	skillContent := `---
name: test-skill
description: A test skill
---

This is a test skill content.
`
	if err := os.WriteFile(filepath.Join(skillDir, SkillFileName), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 加载技能
	err := loader.Load(t.Context())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	skills := loader.List()
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	if skills[0] != "test-skill" {
		t.Errorf("expected skill name 'test-skill', got '%s'", skills[0])
	}

	skill, ok := loader.Get("test-skill")
	if !ok {
		t.Fatal("skill 'test-skill' not found")
	}

	if skill.Name != "test-skill" {
		t.Errorf("expected skill name 'test-skill', got '%s'", skill.Name)
	}
}

func TestLoad_MissingSkillFile(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 创建技能目录但没有 SKILL.md
	skillDir := filepath.Join(loader.skillsDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := loader.Load(t.Context())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 应该跳过没有 SKILL.md 的技能
	skills := loader.List()
	if len(skills) != 0 {
		t.Errorf("expected 0 skills (missing SKILL.md), got %d", len(skills))
	}
}

func TestLoad_NonDirectoryEntries(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 创建一个文件而不是目录（需要先创建 skills 目录）
	if err := os.MkdirAll(loader.skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loader.skillsDir, "not-a-dir"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	err := loader.Load(t.Context())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 应该跳过非目录条目
	skills := loader.List()
	if len(skills) != 0 {
		t.Errorf("expected 0 skills (non-directory), got %d", len(skills))
	}
}

func TestGet(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 创建技能
	skillDir := filepath.Join(loader.skillsDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	skillContent := `---
name: test-skill
description: A test skill
---

This is a test skill content.
`
	if err := os.WriteFile(filepath.Join(skillDir, SkillFileName), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 加载技能
	if err := loader.Load(t.Context()); err != nil {
		t.Fatal(err)
	}

	// 测试获取存在的技能
	skill, ok := loader.Get("test-skill")
	if !ok {
		t.Fatal("skill 'test-skill' not found")
	}

	if skill.Name != "test-skill" {
		t.Errorf("expected skill name 'test-skill', got '%s'", skill.Name)
	}

	// 测试获取不存在的技能
	_, ok = loader.Get("non-existent")
	if ok {
		t.Error("expected false for non-existent skill, got true")
	}
}
