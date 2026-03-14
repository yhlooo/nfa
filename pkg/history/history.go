package history

import (
	"time"
)

// Entry 历史记录条目
type Entry struct {
	TS      int64  `json:"ts"`      // Unix timestamp (nano)
	Content string `json:"content"` // 输入内容
}

// History 历史记录管理器
type History struct {
	entries []Entry
	maxLen  int

	// 导航状态（不持久化）
	navIndex int // 当前浏览位置，-1 表示未在浏览
}

// NewHistory 创建新的历史记录管理器
func NewHistory(maxLen int) *History {
	if maxLen <= 0 {
		maxLen = 100
	}
	return &History{
		entries:  make([]Entry, 0),
		maxLen:   maxLen,
		navIndex: -1,
	}
}

// Add 添加一条新记录
func (h *History) Add(content string) {
	if content == "" {
		return
	}

	entry := Entry{
		TS:      time.Now().UnixNano(),
		Content: content,
	}

	// 添加到末尾（最新）
	h.entries = append(h.entries, entry)

	// 超过最大数量时移除最旧的
	if len(h.entries) > h.maxLen {
		h.entries = h.entries[1:]
	}

	// 重置导航状态
	h.navIndex = -1
}

// Up 向上浏览（获取更早的记录）
// 返回空字符串表示已到达最旧
func (h *History) Up() string {
	if len(h.entries) == 0 {
		return ""
	}

	// 如果未在浏览，从最新的开始
	if h.navIndex == -1 {
		h.navIndex = len(h.entries) - 1
	} else if h.navIndex > 0 {
		h.navIndex--
	}

	return h.entries[h.navIndex].Content
}

// Down 向下浏览（获取更新的记录）
// 返回空字符串表示已到达最新
func (h *History) Down() string {
	if len(h.entries) == 0 || h.navIndex == -1 {
		return ""
	}

	// 如果已经在最新的历史记录，返回空并重置
	if h.navIndex == len(h.entries)-1 {
		h.navIndex = -1
		return ""
	}

	h.navIndex++
	return h.entries[h.navIndex].Content
}

// IsNavigating 返回是否正在浏览历史
func (h *History) IsNavigating() bool {
	return h.navIndex != -1
}

// ResetNav 重置导航状态
func (h *History) ResetNav() {
	h.navIndex = -1
}

// Entries 返回所有历史记录（只读）
func (h *History) Entries() []Entry {
	return h.entries
}

// SetEntries 设置历史记录（用于加载）
func (h *History) SetEntries(entries []Entry) {
	h.entries = entries
	h.navIndex = -1
}

// Count 返回历史记录数量
func (h *History) Count() int {
	return len(h.entries)
}
