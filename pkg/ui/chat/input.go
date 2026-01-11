package chat

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-logr/logr"
)

// NewInputBox 创建输入框
func NewInputBox(ctx context.Context, commands []SelectorOption) *InputBox {
	logger := logr.FromContextOrDiscard(ctx)

	input := textarea.New()
	input.ShowLineNumbers = false
	input.SetWidth(100)
	input.FocusedStyle.Base = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false)
	input.Focus()

	commandSelector := NewSelector(commands, "/", 8, 100)

	return &InputBox{
		logger:          logger,
		input:           input,
		commandSelector: commandSelector,
		rightTipsStyle:  lipgloss.NewStyle().Width(100).AlignHorizontal(lipgloss.Right),
		faintTipsStyle:  lipgloss.NewStyle().Faint(true),
		width:           100,
	}
}

// InputBox 输入框
type InputBox struct {
	logger logr.Logger

	input           textarea.Model
	commandSelector Selector

	rightTipsStyle lipgloss.Style
	faintTipsStyle lipgloss.Style

	multiLine bool
	width     int
	doubleEsc bool
}

// Update 处理更新事件
func (box *InputBox) Update(msg tea.Msg) (*InputBox, tea.Cmd) {
	if !box.Focused() {
		return box, nil
	}

	box.input.SetHeight(10)

	var inputCmd, commandSelectorCmd tea.Cmd
	box.input, inputCmd = box.input.Update(msg)
	box.commandSelector, commandSelectorCmd = box.commandSelector.Update(msg)

	cmds := []tea.Cmd{inputCmd, commandSelectorCmd}

	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		box.width = typedMsg.Width

	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			box.doubleEsc = false

			if val := box.input.Value(); strings.HasPrefix(val, "/") {
				box.commandSelector.SetSearchKey(strings.TrimPrefix(val, "/"))
				box.commandSelector.SetEnabled(true)
			} else {
				box.commandSelector.SetEnabled(false)
			}

		case tea.KeyEnter:
			if box.commandSelector.Enabled() {
				selected := box.commandSelector.Selected()
				if selected != "" {
					box.input.SetValue(selected)
					box.commandSelector.SetEnabled(false)
				}
			}

		case tea.KeyTab:
			if box.commandSelector.Enabled() {
				selected := box.commandSelector.Selected()
				if selected != "" {
					box.input.SetValue(selected)
					box.commandSelector.SetEnabled(false)
				}
			} else {
				box.multiLine = !box.multiLine
			}

		case tea.KeyEsc:
			if box.doubleEsc {
				box.input.Reset()
			}
			box.doubleEsc = !box.doubleEsc
			box.commandSelector.SetEnabled(false)
		default:
			box.doubleEsc = false
		}
	}

	cmds = append(cmds, box.updateComponents()...)

	return box, tea.Batch(cmds...)
}

// updateComponents 根据状态更新组件
func (box *InputBox) updateComponents() []tea.Cmd {
	var cmds []tea.Cmd

	// 设置宽度
	box.input.SetWidth(box.width)
	box.commandSelector.SetWidth(box.width)
	box.rightTipsStyle = box.rightTipsStyle.Width(box.width)

	// 设置多行/单行输入
	inputHeight := 0
	if box.multiLine {
		inputHeight = 10
	} else {
		inputHeight = box.input.LineCount()
		if inputHeight > 10 {
			inputHeight = 10
		}
	}
	if inputHeight > 1 {
		box.input.Prompt = ""
		box.input.ShowLineNumbers = true
	} else {
		box.input.Prompt = "  > "
		box.input.ShowLineNumbers = false
	}
	box.input.SetHeight(inputHeight)
	box.input.KeyMap.InsertNewline.SetEnabled(box.multiLine)

	return cmds
}

// View 渲染显示内容
func (box *InputBox) View() string {
	var ret strings.Builder

	ret.WriteString(box.input.View() + "\n")

	if box.commandSelector.Enabled() {
		ret.WriteString(box.commandSelector.View() + "\n")
	} else if box.multiLine {
		ret.WriteString(box.rightTipsStyle.Render("MULTILINE MODE " + box.faintTipsStyle.Render("(tab to toggle)")))
	}

	return ret.String()
}

// Focused 返回是否获得焦点
func (box *InputBox) Focused() bool {
	return box.input.Focused()
}

// MultiLineMode 返回是否处于多行编辑模式
func (box *InputBox) MultiLineMode() bool {
	return box.multiLine
}

// Value 返回输入内容
func (box *InputBox) Value() string {
	return box.input.Value()
}

// Blur 移除焦点
//
//goland:noinspection GoMixedReceiverTypes
func (box *InputBox) Blur() {
	box.input.Blur()
}

// Focus 获得焦点
//
//goland:noinspection GoMixedReceiverTypes
func (box *InputBox) Focus() tea.Cmd {
	return box.input.Focus()
}

// Reset 重置
//
//goland:noinspection GoMixedReceiverTypes
func (box *InputBox) Reset() {
	box.input.Reset()
	box.commandSelector.SetEnabled(false)
	box.multiLine = false
}

// NewSelector 创建选择器
func NewSelector(items []SelectorOption, suggestionPrefix string, height, width int) Selector {
	s := Selector{
		SuggestionPrefix: suggestionPrefix,
		ShowDescription:  true,
		NamePadding:      4,
		items:            items,
		table: table.New(
			table.WithFocused(true),
			table.WithStyles(table.Styles{
				Header:   lipgloss.NewStyle(),
				Cell:     lipgloss.NewStyle(),
				Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")),
			}),
		),
		height: height,
		width:  width,
	}
	s.RenewTable()
	return s
}

// Selector 选择器
type Selector struct {
	SuggestionPrefix string
	ShowDescription  bool
	NamePadding      int

	items []SelectorOption
	table table.Model

	enabled   bool
	searchKey string
	height    int
	width     int
}

// SelectorOption 选择器选项
type SelectorOption struct {
	Name        string
	Description string
}

// Update 处理更新事件
func (s Selector) Update(msg tea.Msg) (Selector, tea.Cmd) {
	if !s.enabled {
		return s, nil
	}
	var cmd tea.Cmd
	s.table, cmd = s.table.Update(msg)
	return s, cmd
}

// View 渲染显示内容
func (s Selector) View() string {
	if !s.enabled {
		return ""
	}

	view := s.table.View()
	divided := strings.SplitN(view, "\n", 2)
	if len(divided) != 2 {
		return view
	}

	return divided[1]
}

// Enabled 返回是否可用
func (s Selector) Enabled() bool {
	return s.enabled
}

// Selected 返回选中内容
func (s Selector) Selected() string {
	selected := s.table.SelectedRow()
	if selected == nil || len(selected) < 1 {
		return ""
	}
	return selected[0]
}

// SetWidth 设置宽度
//
//goland:noinspection GoMixedReceiverTypes
func (s *Selector) SetWidth(width int) {
	if width == s.width {
		return
	}
	s.width = width
	s.RenewTable()
}

// SetSearchKey 设置搜索键
//
//goland:noinspection GoMixedReceiverTypes
func (s *Selector) SetSearchKey(key string) {
	if s.searchKey == key {
		return
	}
	s.searchKey = key
	s.RenewTable()
}

// SetEnabled 设置是否启用
//
//goland:noinspection GoMixedReceiverTypes
func (s *Selector) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// RenewTable 更新表格
//
//goland:noinspection GoMixedReceiverTypes
func (s *Selector) RenewTable() {
	// 行
	maxNameLen := 0
	rows := make([]table.Row, 0, len(s.items))
	for _, item := range s.items {
		if strings.Contains(item.Name, s.searchKey) {
			row := table.Row{s.SuggestionPrefix + item.Name}
			if s.ShowDescription {
				row = append(row, item.Description)
			}
			if len(item.Name) > maxNameLen {
				maxNameLen = len(item.Name)
			}
			rows = append(rows, row)
		}
	}

	cols := []table.Column{{Title: "Name", Width: maxNameLen + s.NamePadding}}
	if s.ShowDescription {
		cols = append(cols, table.Column{Title: "Description", Width: s.width - maxNameLen - s.NamePadding})
	}

	s.table.SetColumns(cols)
	s.table.SetRows(rows)

	height := len(rows)
	if height > s.height {
		height = s.height
	}
	s.table.SetHeight(height + 1)

	s.table.GotoTop()
}
