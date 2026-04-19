package chat

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	i18nutil "github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/models"
)

// ModelSelector 模型选择器
type ModelSelector struct {
	availableModels []models.ModelConfig

	modelType ModelType
	cursor    int // 当前选中的索引

	selectedModels models.Models
	ctx            context.Context
}

// NewModelSelector 创建模型选择器
func NewModelSelector() *ModelSelector {
	s := &ModelSelector{
		modelType: ModelTypePrimary,
		cursor:    0,
	}
	s.syncCursor()

	return s
}

// Init 初始化
func (s *ModelSelector) Init() tea.Cmd {
	return nil
}

// Update 处理更新事件
func (s *ModelSelector) Update(msg tea.Msg) (*ModelSelector, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyUp, tea.KeyShiftTab:
			if s.cursor > 0 {
				s.cursor--
			}
		case tea.KeyDown, tea.KeyTab:
			if s.cursor < len(s.availableModels)-1 {
				s.cursor++
			}
		case tea.KeyEnter:
			s.syncCurrentModels()
		default:
		}
	}

	return s, nil
}

// View 渲染显示内容
func (s *ModelSelector) View() string {
	var b strings.Builder

	// 标题
	title := i18nutil.LocalizeContext(s.ctx, &i18n.LocalizeConfig{
		DefaultMessage: MsgSelectModel,
		TemplateData:   map[string]any{"Type": s.modelType},
	})
	b.WriteString(title + "\n")
	b.WriteString("\n")

	// 模型列表
	for i, item := range s.availableModels {
		cursor := " "
		if i == s.cursor {
			cursor = "❯"
		}

		// 格式: " ❯ 1. ollama/llama3.2"
		line := fmt.Sprintf(" %s %d. %s", cursor, i+1, item.Name)

		b.WriteString(line + "\n")
	}

	return b.String()
}

// GetSelectedModels 返回更新后的模型选择
func (s *ModelSelector) GetSelectedModels() (ModelType, string, models.Models) {
	return s.modelType, s.availableModels[s.cursor].Name, s.selectedModels
}

// SetCurrentModels 设置当前模型
func (s *ModelSelector) SetCurrentModels(curModels models.Models) {
	s.selectedModels = curModels
	s.syncCursor()
}

// SetAvailableModels 设置可用模型列表
func (s *ModelSelector) SetAvailableModels(availableModels []models.ModelConfig) {
	s.availableModels = availableModels
	s.syncCursor()
}

// SetModelType 设置选择的模型类型
func (s *ModelSelector) SetModelType(t ModelType) {
	s.modelType = t
	s.syncCursor()
}

// SetContext 设置上下文
func (s *ModelSelector) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// syncCursor 同步指针
func (s *ModelSelector) syncCursor() {
	currentName := ""
	switch s.modelType {
	case ModelTypePrimary:
		currentName = s.selectedModels.Primary
	case ModelTypeLight:
		currentName = s.selectedModels.Light
	case ModelTypeVision:
		currentName = s.selectedModels.Vision
	}

	cursor := 0
	for i, item := range s.availableModels {
		if item.Name == currentName {
			cursor = i
			break
		}
	}

	s.cursor = cursor

	return
}

// syncCurrentModels 同步当前模型
func (s *ModelSelector) syncCurrentModels() {
	if s.cursor < 0 || s.cursor >= len(s.availableModels) {
		return
	}
	modelName := s.availableModels[s.cursor].Name

	switch s.modelType {
	case ModelTypePrimary:
		s.selectedModels.Primary = modelName
	case ModelTypeLight:
		s.selectedModels.Light = modelName
	case ModelTypeVision:
		s.selectedModels.Vision = modelName
	}
}
