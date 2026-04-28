package memory

import (
	"os"
	"sync"
)

// Memory 长期记忆
type Memory struct {
	mu       sync.RWMutex
	filePath string
	content  string
}

// Load 从文件加载记忆，如果文件不存在则返回空记忆
func Load(filePath string) (*Memory, error) {
	m := &Memory{
		filePath: filePath,
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}
	m.content = string(data)
	return m, nil
}

// Content 返回记忆内容（线程安全）
func (m *Memory) Content() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.content
}

// Update 更新记忆内容并写入文件
func (m *Memory) Update(newContent string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.content = newContent
	return os.WriteFile(m.filePath, []byte(newContent), 0o644)
}
