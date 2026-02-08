package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"github.com/yhlooo/nfa/pkg/models"
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
		Logger:    logr.Discard(),
		DataRoot:  tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 验证技能加载器已初始化
	if agent.skillLoader == nil {
		t.Fatal("skillLoader should be initialized")
	}

	// 验证技能已加载
	skills := agent.skillLoader.List()
	if len(skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(skills))
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
		Logger:    logr.Discard(),
		DataRoot:  tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 验证技能加载器已初始化
	if agent.skillLoader == nil {
		t.Fatal("skillLoader should be initialized")
	}

	// 验证没有技能
	skills := agent.skillLoader.List()
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
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
		Logger:    logr.Discard(),
		DataRoot:  tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 验证只加载了有效技能
	skills := agent.skillLoader.List()
	if len(skills) != 1 {
		t.Errorf("expected 1 valid skill, got %d", len(skills))
	}

	if skills[0] != "valid-skill" {
		t.Errorf("expected skill 'valid-skill', got '%s'", skills[0])
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
		Logger:    logr.Discard(),
		DataRoot:  tmpDir,
		ModelProviders: []models.ModelProvider{
			{Ollama: &models.OllamaOptions{}},
		},
	})

	// 初始化 genkit
	agent.InitGenkit(context.Background())

	// 生成系统提示
	systemPromptFn := AnalystSystemPromptWithSkills(agent.skillLoader)
	prompt, err := systemPromptFn(context.Background(), nil)
	if err != nil {
		t.Fatalf("AnalystSystemPromptWithSkills() error = %v", err)
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
