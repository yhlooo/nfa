package tty

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NewInputBox 创建输入框
func NewInputBox(commands []SelectorOption) InputBox {

	input := textarea.New()
	input.Prompt = "> "
	input.ShowLineNumbers = false
	input.SetWidth(100)
	input.SetHeight(1)
	input.Focus()

	commandSelector := NewSelector(commands, "/", 8, 100)

	return InputBox{
		input:           input,
		commandSelector: commandSelector,
		enabled:         true,
		width:           100,
	}
}

// InputBox 输入框
type InputBox struct {
	input           textarea.Model
	commandSelector Selector

	enabled   bool
	width     int
	doubleEsc bool
}

// Update 处理更新事件
func (box InputBox) Update(msg tea.Msg) (InputBox, tea.Cmd) {
	if !box.enabled {
		return box, nil
	}

	var inputCmd, commandSelectorCmd tea.Cmd
	box.input, inputCmd = box.input.Update(msg)
	box.commandSelector, commandSelectorCmd = box.commandSelector.Update(msg)

	cmds := []tea.Cmd{inputCmd, commandSelectorCmd}

	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		box.width = typedMsg.Width
		box.input.SetWidth(typedMsg.Width)
		box.commandSelector.SetWidth(typedMsg.Width)

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

		case tea.KeyEnter, tea.KeyTab:
			if box.commandSelector.Enabled() {
				selected := box.commandSelector.Selected()
				if selected != "" {
					box.input.SetValue(selected)
					box.commandSelector.SetEnabled(false)
				}
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

	return box, tea.Batch(cmds...)
}

// View 渲染显示内容
func (box InputBox) View() string {
	var ret strings.Builder
	ret.WriteString(strings.Repeat("─", box.width) + "\n")
	ret.WriteString(box.input.View() + "\n")
	ret.WriteString(strings.Repeat("─", box.width) + "\n")

	if box.commandSelector.Enabled() {
		ret.WriteString(box.commandSelector.View() + "\n")
	}

	return ret.String()
}

// Enabled 返回是否启用
func (box InputBox) Enabled() bool {
	return box.enabled
}

// Value 返回输入内容
func (box InputBox) Value() string {
	return box.input.Value()
}

// SetEnabled 设置是否启用
//
//goland:noinspection GoMixedReceiverTypes
func (box *InputBox) SetEnabled(enabled bool) {
	box.enabled = enabled
	box.commandSelector.SetEnabled(enabled)
}

// Reset 重置
//
//goland:noinspection GoMixedReceiverTypes
func (box *InputBox) Reset() {
	box.input.Reset()
	box.commandSelector.SetEnabled(false)
}

// NewSelector 创建选择器
func NewSelector(items []SelectorOption, suggestionPrefix string, height, width int) Selector {
	s := Selector{
		SuggestionPrefix: suggestionPrefix,
		ShowDescription:  true,
		NamePadding:      4,
		items:            items,
		table: table.New(
			table.WithHeight(height+1),
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
	s.table.GotoTop()
}
