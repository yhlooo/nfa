package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-logr/logr"
)

// SkillLoader 技能加载器
type SkillLoader struct {
	lock sync.RWMutex

	skillsDir  string
	skillRefs  map[string]SkillRef
	skillNames []string
}

// SkillRef 技能
type SkillRef struct {
	Meta SkillMeta
	Path string
}

// NewSkillLoader 创建技能加载器
func NewSkillLoader(skillsDir string) *SkillLoader {
	return &SkillLoader{
		skillsDir: skillsDir,
	}
}

// LoadMeta 加载所有技能元信息
func (sl *SkillLoader) LoadMeta(ctx context.Context) error {
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
	sl.skillRefs = make(map[string]SkillRef)
	sl.skillNames = nil

	// 加载每个技能
	for _, entry := range entries {
		if !entry.IsDir() {
			// 跳过非目录文件
			continue
		}

		skillName := entry.Name()
		skillPath := filepath.Join(sl.skillsDir, skillName)

		// 验证 SKILL.md 格式
		s, err := ReadSkill(skillPath)
		if err != nil {
			logger.Info(fmt.Sprintf("WARN read skill %q from %q error, skipped: %s", skillName, skillPath, err))
			continue
		}

		// 添加到技能列表
		sl.skillRefs[skillName] = SkillRef{
			Meta: s.Meta,
			Path: skillPath,
		}
		sl.skillNames = append(sl.skillNames, skillName)
		logger.Info(fmt.Sprintf("loaded skill: %s", skillName))
	}

	logger.Info(fmt.Sprintf("loaded %d skills", len(sl.skillRefs)))
	return nil
}

// Get 获取技能
func (sl *SkillLoader) Get(name string) (*Skill, error) {
	sl.lock.RLock()
	defer sl.lock.RUnlock()

	ref, ok := sl.skillRefs[name]
	if !ok {
		return nil, fmt.Errorf("skill %q not found", name)
	}

	skill, err := ReadSkill(ref.Path)
	if err != nil {
		return nil, fmt.Errorf("read skill %q from %q error: %w", name, ref.Path, err)
	}

	return skill, nil
}

// ListMeta 列出所有技能元数据
func (sl *SkillLoader) ListMeta() []SkillMeta {
	sl.lock.RLock()
	defer sl.lock.RUnlock()

	ret := make([]SkillMeta, len(sl.skillNames))
	for i, name := range sl.skillNames {
		ret[i] = sl.skillRefs[name].Meta
	}

	return ret
}
