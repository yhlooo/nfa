package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"github.com/yhlooo/nfa/pkg/models"
	"github.com/yhlooo/nfa/pkg/skills"
)

func TestNFAAgent_WithSkills(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试技能目录
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建技能 1
	skill1Dir := filepath.Join(skillsDir, "get-price")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatal(err)
	}
	skill1Content := `---
name: get-price
description: Get asset prices
---

1. Confirm asset code
2. Query recent prices
`
	if err := os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建技能 2
	skill2Dir := filepath.Join(skillsDir, "analyze-trend")
	if err := os.MkdirAll(skill2Dir, 0755); err != nil {
		t.Fatal(err)
	}
	skill2Content := `---
name: analyze-trend
description: Analyze market trends
---

1. Get historical data
2. Calculate indicators
`
	if err := os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 agent
	agent := NewNFA(Options{
		Logger:   logr.Discard(),
		DataRoot: tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化技能加载器
	agent.skillLoader = skills.NewSkillLoader(filepath.Join(tmpDir, "skills"))

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 初始化技能加载器
	if err := agent.skillLoader.LoadMeta(context.Background()); err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 验证技能加载器已初始化
	if agent.skillLoader == nil {
		t.Fatal("skillLoader should be initialized")
	}

	// 验证技能已加载
	skills := agent.skillLoader.ListMeta()
	// 2 个用户技能 + 1 个内置技能
	if len(skills) < 3 {
		t.Errorf("expected at least 3 skills (2 user + builtin), got %d", len(skills))
	}

	// 验证工具已注册
	if len(agent.availableTools) == 0 {
		t.Fatal("no tools registered")
	}

	// 验证 Skill 工具已注册
	skillToolFound := false
	for _, tool := range agent.availableTools {
		if tool.Name() == "Skill" {
			skillToolFound = true
			break
		}
	}
	if !skillToolFound {
		t.Error("Skill tool should be registered")
	}
}

func TestNFAAgent_EmptySkillsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建空的技能目录
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建 agent
	agent := NewNFA(Options{
		Logger:   logr.Discard(),
		DataRoot: tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化技能加载器
	agent.skillLoader = skills.NewSkillLoader(filepath.Join(tmpDir, "skills"))

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 初始化技能加载器
	if err := agent.skillLoader.LoadMeta(context.Background()); err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 验证技能加载器已初始化
	if agent.skillLoader == nil {
		t.Fatal("skillLoader should be initialized")
	}

	// 验证至少有内置技能
	skillList := agent.skillLoader.ListMeta()
	if len(skillList) == 0 {
		t.Error("expected at least builtin skill, got 0")
	}

	// 验证 Skill 工具仍被注册
	skillToolFound := false
	for _, tool := range agent.availableTools {
		if tool.Name() == "Skill" {
			skillToolFound = true
			break
		}
	}
	if !skillToolFound {
		t.Error("Skill tool should still be registered even with no skills")
	}
}

func TestNFAAgent_InvalidSkill(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建技能目录
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建一个有效技能
	validSkillDir := filepath.Join(skillsDir, "valid-skill")
	if err := os.MkdirAll(validSkillDir, 0755); err != nil {
		t.Fatal(err)
	}
	validSkillContent := `---
name: valid-skill
description: A valid skill
---

Valid skill content.
`
	if err := os.WriteFile(filepath.Join(validSkillDir, "SKILL.md"), []byte(validSkillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建一个无效技能（没有 SKILL.md）
	invalidSkillDir := filepath.Join(skillsDir, "invalid-skill")
	if err := os.MkdirAll(invalidSkillDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建 agent
	agent := NewNFA(Options{
		Logger:   logr.Discard(),
		DataRoot: tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化技能加载器
	agent.skillLoader = skills.NewSkillLoader(filepath.Join(tmpDir, "skills"))

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 初始化技能加载器
	if err := agent.skillLoader.LoadMeta(context.Background()); err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 验证加载了有效技能和内置技能
	skillList := agent.skillLoader.ListMeta()
	if len(skillList) < 2 {
		t.Errorf("expected at least 2 skills (valid + builtin), got %d", len(skillList))
	}

	// 验证包含有效技能
	found := false
	for _, s := range skillList {
		if s.Name == "valid-skill" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find 'valid-skill'")
	}
}

func TestAnalystSystemPromptWithSkills(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试技能目录
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建技能
	skillDir := filepath.Join(skillsDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	skillContent := `---
name: test-skill
description: A test skill
---

Test content.
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 agent
	agent := NewNFA(Options{
		Logger:   logr.Discard(),
		DataRoot: tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化技能加载器
	agent.skillLoader = skills.NewSkillLoader(filepath.Join(tmpDir, "skills"))

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 初始化技能加载器
	if err := agent.skillLoader.LoadMeta(context.Background()); err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 生成系统提示
	systemPromptFn := AnalystSystemPrompt(agent.skillLoader)
	prompt, err := systemPromptFn(context.Background(), nil)
	if err != nil {
		t.Fatalf("AnalystSystemPrompt() error = %v", err)
	}

	// 验证提示包含技能信息
	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}

	// 验证包含技能名称
	if len(prompt) < 10 {
		t.Error("prompt should be longer")
	}
}

// TestNFAAgent_WithBuiltinSkills 测试内置技能加载
func TestNFAAgent_WithBuiltinSkills(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建 agent（不创建用户技能目录）
	agent := NewNFA(Options{
		Logger:   logr.Discard(),
		DataRoot: tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化技能加载器
	agent.skillLoader = skills.NewSkillLoader(filepath.Join(tmpDir, "skills"))

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 加载技能（包含内置技能）
	if err := agent.skillLoader.LoadMeta(context.Background()); err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 验证内置技能已加载
	skillList := agent.skillLoader.ListMeta()
	if len(skillList) == 0 {
		t.Error("expected at least 1 builtin skill, got 0")
	}

	// 验证内置技能 short-term-trend-forecast 存在
	found := false
	for _, skill := range skillList {
		if skill.Name == "short-term-trend-forecast" {
			found = true
			if skill.Source != skills.SkillSourceBuiltin {
				t.Errorf("expected source '%s', got '%s'", skills.SkillSourceBuiltin, skill.Source)
			}
			break
		}
	}
	if !found {
		t.Error("builtin skill 'short-term-trend-forecast' not found")
	}

	// 验证 Skill 工具已注册
	if len(agent.availableTools) == 0 {
		t.Fatal("no tools registered")
	}
}

// TestNFAAgent_BuiltinAndUserSkills 测试内置技能和用户技能共存
func TestNFAAgent_BuiltinAndUserSkills(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建用户技能目录
	skillsDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建用户技能
	userSkillDir := filepath.Join(skillsDir, "user-custom-skill")
	if err := os.MkdirAll(userSkillDir, 0755); err != nil {
		t.Fatal(err)
	}
	userSkillContent := `---
name: user-custom-skill
description: A custom user skill
---
Custom user skill content.`
	if err := os.WriteFile(filepath.Join(userSkillDir, "SKILL.md"), []byte(userSkillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 agent
	agent := NewNFA(Options{
		Logger:   logr.Discard(),
		DataRoot: tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化技能加载器
	agent.skillLoader = skills.NewSkillLoader(filepath.Join(tmpDir, "skills"))

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 加载技能
	if err := agent.skillLoader.LoadMeta(context.Background()); err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}

	// 验证两个技能都存在
	skillList := agent.skillLoader.ListMeta()
	if len(skillList) < 2 {
		t.Errorf("expected at least 2 skills (builtin + user), got %d", len(skillList))
	}

	hasBuiltin := false
	hasUser := false
	for _, skill := range skillList {
		if skill.Name == "short-term-trend-forecast" && skill.Source == skills.SkillSourceBuiltin {
			hasBuiltin = true
		}
		if skill.Name == "user-custom-skill" && skill.Source == skills.SkillSourceLocal {
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
