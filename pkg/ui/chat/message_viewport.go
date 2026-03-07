package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/coder/acp-go-sdk"
	"github.com/go-logr/logr"
)

// NewMessageViewport 创建消息视窗
func NewMessageViewport(ctx context.Context) (MessageViewport, error) {
	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		return MessageViewport{}, fmt.Errorf("new markdown renderer error: %w", err)
	}
	return MessageViewport{
		viewStyle:  lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderLeft(true),
		mdRenderer: r,
		logger:     logr.FromContextOrDiscard(ctx),
	}, nil
}

// MessageViewport 消息视窗
type MessageViewport struct {
	agentProcessing int
	messages        MessagesList
	viewStyle       lipgloss.Style
	mdRenderer      *glamour.TermRenderer
	logger          logr.Logger
}

// AgentProcessing 返回是否 Agent 处理中
func (vp MessageViewport) AgentProcessing() bool {
	return vp.agentProcessing > 0
}

// Update 处理更新事件
func (vp MessageViewport) Update(msg tea.Msg) (MessageViewport, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		vp.viewStyle = vp.viewStyle.Width(typedMsg.Width)

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
func (vp MessageViewport) View() string {
	if len(vp.messages) == 0 {
		return ""
	}
	return vp.viewStyle.Render(vp.viewMessages(vp.messages, true))
}

// viewMessages 渲染消息
func (vp MessageViewport) viewMessages(messages MessagesList, renderMD bool) string {
	var ret strings.Builder
	for _, msg := range messages {
		switch msg.Type {
		case MessageTypeUser:
			if strings.HasPrefix(msg.Text, "/") {
				ret.WriteString("👉 \033[1;32m" + withIndent(msg.Text, 2) + "\033[0m\n")
			} else {
				ret.WriteString("☝️ \033[1;32m" + withIndent(msg.Text, 2) + "\033[0m\n")
			}
		case MessageTypeAgent:
			if renderMD {
				ret.WriteString(vp.renderMarkdown(msg.Text) + "\n")
			} else {
				ret.WriteString(msg.Text + "\n")
			}
		case MessageTypeAgentThought:
			content := msg.Text
			//if renderMD {
			//	content = vp.renderMarkdown(msg.Text)
			//}
			ret.WriteString("🧠 \033[2m" + withIndent(content, 2) + "\033[0m\n")
		case MessageTypeToolCall:
			ret.WriteString("🔧 \033[34m" + withIndent(msg.Text, 2) + "\033[0m\n")
		case MessageTypeToolCallUpdate:
			ret.WriteString("  \033[34m" + withIndent(msg.Text, 2) + "\033[0m\n")
		case MessageTypeError:
			ret.WriteString("❌ \033[31m" + withIndent(msg.Text, 2) + "\033[0m\n")
		case MessageTypeUnknown:
			ret.WriteString(msg.Text + "\n")
		}
	}
	return strings.TrimRight(ret.String(), "\n")
}

// Flush 将缓存的消息刷到屏幕上
//
//goland:noinspection GoMixedReceiverTypes
func (vp *MessageViewport) Flush() tea.Cmd {
	n := len(vp.messages) - 1
	if vp.agentProcessing == 0 {
		n = len(vp.messages)
	}
	if n <= 0 {
		return nil
	}

	content := vp.viewMessages(vp.messages[:n], true)
	vp.messages = vp.messages[n:]

	return tea.Println(content)
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
		"ToolCall: %s \033[2m%s\033[22m",
		msg.Title,
		withIndent(renderAgentToolCallContent(msg.Content), 11+len(msg.Title)),
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
		status = "\033[31m" + status + "\033[39m"
	}
	return fmt.Sprintf(
		"          %s \033[2m%s\033[22m",
		status,
		withIndent(renderAgentToolCallContent(msg.Content), 11+len(status)),
	)
}

// withIndent 返回带缩进的文本
func withIndent(content string, indent int) string {
	indentStr := strings.Repeat(" ", indent)
	return strings.ReplaceAll(content, "\n", "\n"+indentStr)
}

// renderMarkdown 使用 glamour 渲染 Markdown 文本
// 如果渲染器未初始化或渲染失败，返回原始文本
func (vp MessageViewport) renderMarkdown(text string) string {
	rendered, err := vp.mdRenderer.Render(text)
	if err != nil {
		return text
	}
	return strings.TrimRight(rendered, "\n")
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
