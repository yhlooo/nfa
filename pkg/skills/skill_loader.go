package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-logr/logr"
)

const (
	// SkillsDirName skills 目录名
	SkillsDirName = ".nfa/skills"
	// SkillFileName skill 文件名
	SkillFileName = "SKILL.md"
)

// SkillLoader 技能加载器
type SkillLoader struct {
	lock sync.RWMutex

	logger     logr.Logger
	skillsDir  string
	skills     map[string]*Skill
	skillNames []string
}

// Skill 技能
type Skill struct {
	Name        string
	Description string
	Path        string
}

// NewSkillLoader 创建技能加载器（使用 homeDir）
func NewSkillLoader(logger logr.Logger, homeDir string) *SkillLoader {
	return NewSkillLoaderWithDir(logger, filepath.Join(homeDir, SkillsDirName))
}

// NewSkillLoaderWithDir 创建技能加载器（直接指定 skills 目录）
func NewSkillLoaderWithDir(logger logr.Logger, skillsDir string) *SkillLoader {
	return &SkillLoader{
		logger:    logger.WithName("skill_loader"),
		skillsDir: skillsDir,
		skills:    make(map[string]*Skill),
	}
}

// Load 加载所有技能
func (sl *SkillLoader) Load() error {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	// 创建 skills 目录（如果不存在）
	if err := os.MkdirAll(sl.skillsDir, 0755); err != nil {
		return fmt.Errorf("create skills directory error: %w", err)
	}

	// 扫描技能目录
	entries, err := os.ReadDir(sl.skillsDir)
	if err != nil {
		return fmt.Errorf("read skills directory error: %w", err)
	}

	// 清空现有技能
	sl.skills = make(map[string]*Skill)
	sl.skillNames = nil

	// 加载每个技能
	for _, entry := range entries {
		if !entry.IsDir() {
			// 跳过非目录文件
			continue
		}

		skillName := entry.Name()
		skillPath := filepath.Join(sl.skillsDir, skillName, SkillFileName)

		// 检查 SKILL.md 是否存在
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			sl.logger.Info(fmt.Sprintf("skill '%s' does not have %s file, skipping", skillName, SkillFileName))
			continue
		}

		// 验证 SKILL.md 格式
		_, err := ParseSkillFile(skillPath)
		if err != nil {
			sl.logger.Info(fmt.Sprintf("skill '%s' has invalid format (%v), skipping", skillName, err))
			continue
		}

		// 添加到技能列表
		sl.skills[skillName] = &Skill{
			Name: skillName,
			Path: skillPath,
		}
		sl.skillNames = append(sl.skillNames, skillName)
		sl.logger.Info(fmt.Sprintf("loaded skill: %s", skillName))
	}

	sl.logger.Info(fmt.Sprintf("loaded %d skills", len(sl.skills)))
	return nil
}

// Discover 发现技能（返回技能名称列表）
func (sl *SkillLoader) Discover() ([]string, error) {
	// 创建 skills 目录（如果不存在）
	if err := os.MkdirAll(sl.skillsDir, 0755); err != nil {
		return nil, fmt.Errorf("create skills directory error: %w", err)
	}

	// 扫描技能目录
	entries, err := os.ReadDir(sl.skillsDir)
	if err != nil {
		return nil, fmt.Errorf("read skills directory error: %w", err)
	}

	var skillNames []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillPath := filepath.Join(sl.skillsDir, skillName, SkillFileName)

		// 检查 SKILL.md 是否存在
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			continue
		}

		skillNames = append(skillNames, skillName)
	}

	return skillNames, nil
}

// Get 获取技能
func (sl *SkillLoader) Get(name string) (*Skill, bool) {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	skill, ok := sl.skills[name]
	return skill, ok
}

// List 列出所有技能名称
func (sl *SkillLoader) List() []string {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	return sl.skillNames
}

// GetAll 获取所有技能
func (sl *SkillLoader) GetAll() map[string]*Skill {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	ret := make(map[string]*Skill, len(sl.skills))
	for k, v := range sl.skills {
		ret[k] = v
	}
	return ret
}

// SkillsDir 获取技能目录路径
func (sl *SkillLoader) SkillsDir() string {
	return sl.skillsDir
}
