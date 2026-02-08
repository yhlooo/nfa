package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-logr/logr"
)

const (
	// SkillFileName skill 文件名
	SkillFileName = "SKILL.md"
)

// SkillLoader 技能加载器
type SkillLoader struct {
	lock sync.RWMutex

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

// NewSkillLoader 创建技能加载器
func NewSkillLoader(skillsDir string) *SkillLoader {
	return &SkillLoader{
		skillsDir: skillsDir,
		skills:    make(map[string]*Skill),
	}
}

// Load 加载所有技能
func (sl *SkillLoader) Load(ctx context.Context) error {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	logger := logr.FromContextOrDiscard(ctx)

	// 扫描技能目录
	entries, err := os.ReadDir(sl.skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
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
			logger.Info(fmt.Sprintf("skill '%s' does not have %s file, skipping", skillName, SkillFileName))
			continue
		}

		// 验证 SKILL.md 格式
		_, err := ParseSkillFile(skillPath)
		if err != nil {
			logger.Info(fmt.Sprintf("skill '%s' has invalid format (%v), skipping", skillName, err))
			continue
		}

		// 添加到技能列表
		sl.skills[skillName] = &Skill{
			Name: skillName,
			Path: skillPath,
		}
		sl.skillNames = append(sl.skillNames, skillName)
		logger.Info(fmt.Sprintf("loaded skill: %s", skillName))
	}

	logger.Info(fmt.Sprintf("loaded %d skills", len(sl.skills)))
	return nil
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
