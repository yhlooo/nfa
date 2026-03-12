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

	err := loader.LoadMeta(t.Context())
	if err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	skills := loader.ListMeta()
	// 现在会加载内置技能
	if len(skills) == 0 {
		t.Error("expected at least 1 builtin skill, got 0")
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
	err := loader.LoadMeta(t.Context())
	if err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	skills := loader.ListMeta()
	// 用户技能 + 内置技能
	if len(skills) < 2 {
		t.Fatalf("expected at least 2 skills (user + builtin), got %d", len(skills))
	}

	// 查找用户技能
	found := false
	for _, s := range skills {
		if s.Name == "test-skill" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find 'test-skill'")
	}

	skill, err := loader.Get("test-skill")
	if err != nil {
		t.Fatalf("skill 'test-skill' not found: %v", err)
	}

	if skill.Meta.Name != "test-skill" {
		t.Errorf("expected skill name 'test-skill', got '%s'", skill.Meta.Name)
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

	err := loader.LoadMeta(t.Context())
	if err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 应该跳过没有 SKILL.md 的技能，但内置技能仍会被加载
	skills := loader.ListMeta()
	if len(skills) == 0 {
		t.Error("expected at least builtin skill, got 0")
	}
	// 验证用户创建的无效技能没有被加载
	for _, s := range skills {
		if s.Name == "test-skill" {
			t.Error("invalid skill 'test-skill' should not be loaded")
		}
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

	err := loader.LoadMeta(t.Context())
	if err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 应该跳过非目录条目，但内置技能仍会被加载
	skills := loader.ListMeta()
	if len(skills) == 0 {
		t.Error("expected at least builtin skill, got 0")
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
	if err := loader.LoadMeta(t.Context()); err != nil {
		t.Fatal(err)
	}

	// 测试获取存在的技能
	skill, err := loader.Get("test-skill")
	if err != nil {
		t.Fatalf("skill 'test-skill' not found: %v", err)
	}

	if skill.Meta.Name != "test-skill" {
		t.Errorf("expected skill name 'test-skill', got '%s'", skill.Meta.Name)
	}

	// 测试获取不存在的技能
	_, err = loader.Get("non-existent")
	if err == nil {
		t.Error("expected error for non-existent skill, got nil")
	}
}

// TestLoad_BuiltinSkills 测试仅内置技能可用时的加载行为
func TestLoad_BuiltinSkills(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 加载技能（包含内置技能）
	err := loader.LoadMeta(t.Context())
	if err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	skills := loader.ListMeta()
	if len(skills) == 0 {
		t.Error("expected at least 1 builtin skill, got 0")
	}

	// 验证内置技能 short-term-trend-forecast 存在
	found := false
	for _, skill := range skills {
		if skill.Name == "short-term-trend-forecast" {
			found = true
			if skill.Source != SkillSourceBuiltin {
				t.Errorf("expected source '%s', got '%s'", SkillSourceBuiltin, skill.Source)
			}
			break
		}
	}
	if !found {
		t.Error("builtin skill 'short-term-trend-forecast' not found")
	}
}

// TestLoad_BuiltinAndUserSkills 测试内置技能和用户技能共存时的合并行为
func TestLoad_BuiltinAndUserSkills(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 创建用户技能
	userSkillDir := filepath.Join(loader.skillsDir, "user-skill")
	if err := os.MkdirAll(userSkillDir, 0755); err != nil {
		t.Fatal(err)
	}
	userSkillContent := `---
name: user-skill
description: A user skill
---
This is a user skill content.`
	if err := os.WriteFile(filepath.Join(userSkillDir, SkillFileName), []byte(userSkillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 加载技能
	err := loader.LoadMeta(t.Context())
	if err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	skills := loader.ListMeta()
	if len(skills) < 2 {
		t.Errorf("expected at least 2 skills (builtin + user), got %d", len(skills))
	}

	// 验证两个技能都存在
	hasBuiltin := false
	hasUser := false
	for _, skill := range skills {
		if skill.Name == "short-term-trend-forecast" && skill.Source == SkillSourceBuiltin {
			hasBuiltin = true
		}
		if skill.Name == "user-skill" && skill.Source == SkillSourceLocal {
			hasUser = true
		}
	}
	if !hasBuiltin {
		t.Error("builtin skill not found")
	}
	if !hasUser {
		t.Error("user skill not found")
	}
}

// TestLoad_UserOverridesBuiltin 测试用户技能覆盖同名内置技能的行为
func TestLoad_UserOverridesBuiltin(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 创建与内置技能同名的用户技能
	userSkillDir := filepath.Join(loader.skillsDir, "short-term-trend-forecast")
	if err := os.MkdirAll(userSkillDir, 0755); err != nil {
		t.Fatal(err)
	}
	userSkillContent := `---
name: short-term-trend-forecast
description: User override version
---
This is the user override version.`
	if err := os.WriteFile(filepath.Join(userSkillDir, SkillFileName), []byte(userSkillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 加载技能
	err := loader.LoadMeta(t.Context())
	if err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 验证技能被用户版本覆盖
	skill, err := loader.Get("short-term-trend-forecast")
	if err != nil {
		t.Fatalf("skill 'short-term-trend-forecast' not found: %v", err)
	}

	if skill.Meta.Description != "User override version" {
		t.Errorf("expected user override description, got '%s'", skill.Meta.Description)
	}

	if skill.Meta.Source != SkillSourceLocal {
		t.Errorf("expected source '%s', got '%s'", SkillSourceLocal, skill.Meta.Source)
	}
}

// TestGet_BuiltinSkill 测试从 embed.FS 读取内置技能内容
func TestGet_BuiltinSkill(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	// 加载技能
	err := loader.LoadMeta(t.Context())
	if err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 获取内置技能
	skill, err := loader.Get("short-term-trend-forecast")
	if err != nil {
		t.Fatalf("builtin skill 'short-term-trend-forecast' not found: %v", err)
	}

	// 验证内容
	if skill.Meta.Name != "short-term-trend-forecast" {
		t.Errorf("expected skill name 'short-term-trend-forecast', got '%s'", skill.Meta.Name)
	}

	if skill.Meta.Source != SkillSourceBuiltin {
		t.Errorf("expected source '%s', got '%s'", SkillSourceBuiltin, skill.Meta.Source)
	}

	if skill.Content == "" {
		t.Error("expected skill content to be non-empty")
	}
}

// TestLoad_InvalidBuiltinSkill 测试内置技能解析失败时的错误处理
func TestLoad_InvalidBuiltinSkill(t *testing.T) {
	// 这个测试验证内置技能格式错误时不会导致程序崩溃
	// 由于内置技能是编译时嵌入的，我们无法动态修改它来测试错误情况
	// 这里仅测试加载过程不会失败
	tmpDir := t.TempDir()
	loader := NewSkillLoader(tmpDir)

	err := loader.LoadMeta(t.Context())
	if err != nil {
		t.Fatalf("LoadMeta() should not fail even if builtin skill has issues, got: %v", err)
	}

	// 至少应该有一些内置技能
	skills := loader.ListMeta()
	if len(skills) == 0 {
		t.Log("warning: no builtin skills found")
	}
}
