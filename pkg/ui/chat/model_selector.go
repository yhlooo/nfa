package chat

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yhlooo/nfa/pkg/models"
)

// ModelSelector 模型选择器
type ModelSelector struct {
	availableModels []models.ModelConfig

	modelType ModelType
	cursor    int // 当前选中的索引

	selectedModels models.Models
}

// NewModelSelector 创建模型选择器
func NewModelSelector() *ModelSelector {
	s := &ModelSelector{
		modelType: ModelTypeMain,
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
	title := fmt.Sprintf("Select %s model", s.modelType)
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

		// 如果有描述，添加在后面
		if item.Description != "" {
			desc := item.Description
			if len(desc) > 80 {
				desc = desc[:80] + "..."
			}
			line += fmt.Sprintf(" - %s", desc)
		}

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

// syncCursor 同步指针
func (s *ModelSelector) syncCursor() {
	currentName := ""
	switch s.modelType {
	case ModelTypeMain:
		currentName = s.selectedModels.Main
	case ModelTypeFast:
		currentName = s.selectedModels.Fast
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
	case ModelTypeMain:
		s.selectedModels.Main = modelName
	case ModelTypeFast:
		s.selectedModels.Fast = modelName
	case ModelTypeVision:
		s.selectedModels.Vision = modelName
	}
}
