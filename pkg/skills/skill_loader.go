package skills

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sync"

	"github.com/go-logr/logr"
)

//go:embed builtin
var builtinSkillsFS embed.FS

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

// SkillSource 技能来源
type SkillSource string

const (
	// SkillSourceBuiltin 内置技能来源
	SkillSourceBuiltin SkillSource = "builtin"
	// SkillSourceLocal 用户技能来源
	SkillSourceLocal SkillSource = "local"
)

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

	// 清空现有技能
	sl.skillRefs = make(map[string]SkillRef)
	sl.skillNames = nil

	// 加载内置技能
	if err := sl.loadBuiltinSkills(ctx); err != nil {
		logger.Info(fmt.Sprintf("WARN load builtin skills error: %s", err))
	}

	// 加载用户技能
	if err := sl.loadUserSkills(ctx); err != nil {
		logger.Info(fmt.Sprintf("WARN load user skills error: %s", err))
	}

	logger.Info(fmt.Sprintf("loaded %d skills", len(sl.skillRefs)))
	return nil
}

// loadUserSkills 加载用户技能
func (sl *SkillLoader) loadUserSkills(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	// 扫描用户技能目录
	entries, err := os.ReadDir(sl.skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read skills directory error: %w", err)
	}

	// 加载每个技能
	for _, entry := range entries {
		if !entry.IsDir() {
			// 跳过非目录文件
			continue
		}

		skillName := entry.Name()
		skillPath := filepath.Join(sl.skillsDir, skillName)

		// 读取 SKILL.md 文件
		s, err := ReadSkillFromFile(skillPath)
		if err != nil {
			logger.Info(fmt.Sprintf("WARN read skill %q from %q error, skipped: %s", skillName, skillPath, err))
			continue
		}

		// 添加到技能列表
		sl.skillRefs[skillName] = SkillRef{
			Meta: s.Meta,
			Path: skillPath,
		}

		// 添加到技能列表
		if !slices.Contains(sl.skillNames, skillName) {
			sl.skillNames = append(sl.skillNames, skillName)
		}
		logger.Info(fmt.Sprintf("loaded user skill: %s", skillName))
	}

	return nil
}

// loadBuiltinSkills 加载内置技能
func (sl *SkillLoader) loadBuiltinSkills(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	// 读取 builtin 目录下的子目录
	entries, err := builtinSkillsFS.ReadDir("builtin")
	if err != nil {
		return err
	}

	// 遍历每个技能目录
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillPath := path.Join("builtin", skillName)

		// 读取 SKILL.md 文件
		s, err := ReadSkillFromEmbed(skillPath)
		if err != nil {
			logger.Info(fmt.Sprintf("WARN read builtin skill %q error: %s", skillName, err))
			continue
		}

		// 添加到技能列表
		sl.skillRefs[skillName] = SkillRef{
			Meta: s.Meta,
			Path: skillPath,
		}
		sl.skillNames = append(sl.skillNames, skillName)
		logger.Info(fmt.Sprintf("loaded builtin skill: %s", skillName))
	}

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

	// 根据来源选择读取方式
	var skill *Skill
	var err error
	if ref.Meta.Source == SkillSourceBuiltin {
		skill, err = ReadSkillFromEmbed(ref.Path)
	} else {
		skill, err = ReadSkillFromFile(ref.Path)
	}
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
