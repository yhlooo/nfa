package tty

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coder/acp-go-sdk"
)

// NewMessageViewport 创建消息视窗
func NewMessageViewport() MessageViewport {
	return MessageViewport{
		UserStyle:                 lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")),
		AgentStyle:                lipgloss.NewStyle(),
		AgentThoughtStyle:         lipgloss.NewStyle().Faint(true),
		AgentToolCallStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		AgentToolCallContentStyle: lipgloss.NewStyle().Faint(true),
		ErrorStyle:                lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
	}
}

// MessageViewport 消息视窗
type MessageViewport struct {
	agentProcessing int

	messages MessagesList

	UserStyle                 lipgloss.Style
	AgentStyle                lipgloss.Style
	AgentThoughtStyle         lipgloss.Style
	AgentToolCallStyle        lipgloss.Style
	AgentToolCallContentStyle lipgloss.Style
	ErrorStyle                lipgloss.Style
}

// AgentProcessing 返回是否 Agent 处理中
//
//goland:noinspection GoMixedReceiverTypes
func (vp MessageViewport) AgentProcessing() bool {
	return vp.agentProcessing > 0
}

// Update 处理更新事件
//
//goland:noinspection GoMixedReceiverTypes
func (vp MessageViewport) Update(msg tea.Msg) (MessageViewport, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case acp.PromptRequest:
		// 对话请求
		vp.agentProcessing++
		for _, content := range typedMsg.Prompt {
			vp.messages = vp.messages.Append(MessageItem{Type: MessageTypeUser, Text: renderAgentContent(content)})
		}

	case acp.SessionNotification:
		// 会话变更通知
		switch {
		case typedMsg.Update.UserMessageChunk != nil:
			vp.messages = vp.messages.Append(MessageItem{
				Type: MessageTypeUser,
				Text: renderAgentContent(typedMsg.Update.UserMessageChunk.Content),
			})
		case typedMsg.Update.AgentMessageChunk != nil:
			vp.messages = vp.messages.Append(MessageItem{
				Type: MessageTypeAgent,
				Text: renderAgentContent(typedMsg.Update.AgentMessageChunk.Content),
			})
		case typedMsg.Update.AgentThoughtChunk != nil:
			vp.messages = vp.messages.Append(MessageItem{
				Type: MessageTypeAgentThought,
				Text: renderAgentContent(typedMsg.Update.AgentThoughtChunk.Content),
			})
		case typedMsg.Update.ToolCall != nil:
			vp.messages = vp.messages.Append(MessageItem{
				Type: MessageTypeToolCall,
				Text: vp.renderAgentToolCallStartMessage(typedMsg.Update.ToolCall),
			})
		case typedMsg.Update.ToolCallUpdate != nil:
			vp.messages = vp.messages.Append(MessageItem{
				Type: MessageTypeToolCallUpdate,
				Text: vp.renderAgentToolCallUpdateMessage(typedMsg.Update.ToolCallUpdate),
			})
		default:
			raw, _ := json.Marshal(typedMsg.Update)
			vp.messages = vp.messages.Append(MessageItem{Type: MessageTypeUnknown, Text: string(raw)})
		}

	case acp.PromptResponse:
		// 会话响应
		vp.agentProcessing--
		if typedMsg.StopReason != "" && typedMsg.StopReason != acp.StopReasonEndTurn {
			vp.messages = vp.messages.Append(MessageItem{
				Type: MessageTypeError,
				Text: fmt.Sprintf("stop reason: %s", typedMsg.StopReason),
			})
		}

	case error:
		// 错误
		vp.messages = vp.messages.Append(MessageItem{
			Type: MessageTypeError,
			Text: typedMsg.Error(),
		})
	}

	return vp, nil
}

// View 渲染显示内容
//
//goland:noinspection GoMixedReceiverTypes
func (vp MessageViewport) View() string {
	var ret strings.Builder
	for _, msg := range vp.messages {
		switch msg.Type {
		case MessageTypeUser:
			ret.WriteString("🤔 " + vp.UserStyle.Render(withIndent(msg.Text, 2)) + "\n")
		case MessageTypeAgent:
			ret.WriteString(vp.AgentStyle.Render(msg.Text) + "\n")
		case MessageTypeAgentThought:
			ret.WriteString("🧠 " + vp.AgentThoughtStyle.Render(withIndent(msg.Text, 2)) + "\n")
		case MessageTypeToolCall:
			ret.WriteString("🔧 " + vp.AgentToolCallStyle.Render(withIndent(msg.Text, 2)) + "\n")
		case MessageTypeToolCallUpdate:
			ret.WriteString("  " + vp.AgentToolCallStyle.Render(withIndent(msg.Text, 2)) + "\n")
		case MessageTypeError:
			ret.WriteString("❌ " + vp.ErrorStyle.Render(withIndent(msg.Text, 2)) + "\n")
		case MessageTypeUnknown:
			ret.WriteString(msg.Text + "\n")
		}
	}
	return strings.TrimRight(ret.String(), "\n")
}

// Reset 重置
//
//goland:noinspection GoMixedReceiverTypes
func (vp *MessageViewport) Reset() {
	vp.agentProcessing = 0
	vp.messages = nil
}

// renderAgentToolCallStartMessage 渲染开始调用工具信息
//
//goland:noinspection GoMixedReceiverTypes
func (vp MessageViewport) renderAgentToolCallStartMessage(msg *acp.SessionUpdateToolCall) string {
	return fmt.Sprintf(
		"ToolCall: %s %s",
		msg.Title,
		vp.AgentToolCallContentStyle.Render(withIndent(renderAgentToolCallContent(msg.Content), 11+len(msg.Title))),
	)
}

// renderAgentToolCallUpdateMessage 渲染工具调用更新信息
//
//goland:noinspection GoMixedReceiverTypes
func (vp MessageViewport) renderAgentToolCallUpdateMessage(msg *acp.SessionToolCallUpdate) string {
	status := acp.ToolCallStatus("unknown")
	if msg.Status != nil {
		status = *msg.Status
	}
	if status == acp.ToolCallStatusFailed {
		status = acp.ToolCallStatus(vp.ErrorStyle.Render(string(status)))
	}
	return fmt.Sprintf(
		"          %s %s",
		status,
		vp.AgentToolCallContentStyle.Render(withIndent(renderAgentToolCallContent(msg.Content), 11+len(status))),
	)
}

// withIndent 返回带缩进的文本
func withIndent(content string, indent int) string {
	indentStr := strings.Repeat(" ", indent)
	return strings.ReplaceAll(content, "\n", "\n"+indentStr)
}

// renderAgentToolCallContent 将 Agent 工具调用内容转换为文本
func renderAgentToolCallContent(content []acp.ToolCallContent) string {
	var ret strings.Builder
	for _, item := range content {
		if item.Content == nil {
			continue
		}
		line := renderAgentContent(item.Content.Content)
		if len(line) > 128 {
			line = line[:125] + "..."
		}
		ret.WriteString(line + "\n")
	}
	return strings.TrimRight(ret.String(), "\n")
}

// renderAgentContent 将 Agent 输出内容转换为文本
func renderAgentContent(content acp.ContentBlock) string {
	switch {
	case content.Text != nil:
		return content.Text.Text
	case content.ResourceLink != nil:
		return fmt.Sprintf("[%s](%s)", content.ResourceLink.Name, content.ResourceLink.Uri)
	default:
		raw, _ := json.Marshal(content)
		return string(raw)
	}
}

// MessageItem 消息项
type MessageItem struct {
	Type MessageType
	Text string
}

// MessageType 消息类型
type MessageType int

const (
	MessageTypeUser MessageType = iota
	MessageTypeAgent
	MessageTypeAgentThought
	MessageTypeToolCall
	MessageTypeToolCallUpdate
	MessageTypeError
	MessageTypeUnknown
)

// MessagesList 消息列表
type MessagesList []MessageItem

// Append 追加消息
func (l MessagesList) Append(item MessageItem) MessagesList {
	if len(l) == 0 ||
		!slices.Contains([]MessageType{MessageTypeUser, MessageTypeAgent, MessageTypeAgentThought}, item.Type) ||
		item.Type != l[len(l)-1].Type {
		return append(l, item)
	}

	last := l[len(l)-1]
	last.Text += item.Text
	l[len(l)-1] = last

	return l
}
