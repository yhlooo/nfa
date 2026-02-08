package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
)

func TestNewSkillLoader(t *testing.T) {
	loader := NewSkillLoader(logr.Discard(), "/home/user")
	if loader == nil {
		t.Fatal("NewSkillLoader returned nil")
	}
	expectedPath := "/home/user/.nfa/skills"
	if loader.SkillsDir() != expectedPath {
		t.Errorf("expected skills dir %s, got %s", expectedPath, loader.SkillsDir())
	}
}

func TestLoad_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(logr.Discard(), tmpDir)

	err := loader.Load()
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
	loader := NewSkillLoader(logr.Discard(), tmpDir)

	// 创建技能目录
	skillDir := filepath.Join(loader.SkillsDir(), "test-skill")
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
	err := loader.Load()
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
	loader := NewSkillLoader(logr.Discard(), tmpDir)

	// 创建技能目录但没有 SKILL.md
	skillDir := filepath.Join(loader.SkillsDir(), "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := loader.Load()
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
	loader := NewSkillLoader(logr.Discard(), tmpDir)

	// 创建一个文件而不是目录（需要先创建 skills 目录）
	if err := os.MkdirAll(loader.SkillsDir(), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loader.SkillsDir(), "not-a-dir"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 应该跳过非目录条目
	skills := loader.List()
	if len(skills) != 0 {
		t.Errorf("expected 0 skills (non-directory), got %d", len(skills))
	}
}

func TestDiscover(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(logr.Discard(), tmpDir)

	// 创建多个技能
	for i := 1; i <= 3; i++ {
		skillDir := filepath.Join(loader.SkillsDir(), "skill"+string(rune('0'+i)))
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatal(err)
		}

		content := `---
name: skill` + string(rune('0'+i)) + `
description: Skill ` + string(rune('0'+i)) + `
---

Content for skill ` + string(rune('0'+i)) + `.
`
		if err := os.WriteFile(filepath.Join(skillDir, SkillFileName), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 发现技能
	skills, err := loader.Discover()
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if len(skills) != 3 {
		t.Errorf("expected 3 skills, got %d", len(skills))
	}
}

func TestGet(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(logr.Discard(), tmpDir)

	// 创建技能
	skillDir := filepath.Join(loader.SkillsDir(), "test-skill")
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
	if err := loader.Load(); err != nil {
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

func TestGetAll(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(logr.Discard(), tmpDir)

	// 创建技能
	for _, name := range []string{"skill1", "skill2"} {
		skillDir := filepath.Join(loader.SkillsDir(), name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatal(err)
		}

		skillContent := `---
name: ` + name + `
description: Skill ` + name + `
---

Content for ` + name + `.
`
		if err := os.WriteFile(filepath.Join(skillDir, SkillFileName), []byte(skillContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 加载技能
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	// 获取所有技能
	all := loader.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 skills, got %d", len(all))
	}

	// 验证返回的是副本而不是引用
	all["new-skill"] = &Skill{Name: "new-skill"}
	if _, ok := loader.Get("new-skill"); ok {
		t.Error("GetAll should return a copy, not reference")
	}
}
