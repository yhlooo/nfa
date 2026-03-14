package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadHistory 从指定路径加载历史记录
// 如果文件不存在，返回空的 History
func LoadHistory(path string) (*History, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，返回空的历史记录
			return NewHistory(100), nil
		}
		return nil, fmt.Errorf("read history file error: %w", err)
	}

	var entries []Entry
	if err := json.Unmarshal(content, &entries); err != nil {
		return nil, fmt.Errorf("unmarshal history from json error: %w", err)
	}

	h := NewHistory(100)
	h.SetEntries(entries)
	return h, nil
}

// SaveHistory 将历史记录保存到指定路径
func SaveHistory(path string, h *History) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create history directory error: %w", err)
	}

	content, err := json.MarshalIndent(h.Entries(), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history to json error: %w", err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write history file error: %w", err)
	}

	return nil
}
