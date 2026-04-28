package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const dailyMemoryDirName = "memory"
const dailyMemoryFileExt = ".md"

// DailyMemory 中期记忆，按天文件存储
type DailyMemory struct {
	mu      sync.RWMutex
	dirPath string
}

// NewDailyMemory 创建中期记忆存储
func NewDailyMemory(dataRoot string) *DailyMemory {
	return &DailyMemory{
		dirPath: filepath.Join(dataRoot, dailyMemoryDirName),
	}
}

// LoadRecent 加载最近 N 天的记忆文件，按日期降序拼接（最新在前）
func (dm *DailyMemory) LoadRecent(days int) (string, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	today := time.Now()
	var parts []string

	for i := 0; i < days; i++ {
		date := today.AddDate(0, 0, -i).Format(time.DateOnly)
		filePath := filepath.Join(dm.dirPath, date+dailyMemoryFileExt)
		data, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("read daily memory %s: %w", date, err)
		}
		content := strings.TrimSpace(string(data))
		if content != "" {
			parts = append(parts, content)
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

// UpdateToday 将内容写入当天的记忆文件，自动创建目录
func (dm *DailyMemory) UpdateToday(content string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if err := os.MkdirAll(dm.dirPath, 0o755); err != nil {
		return fmt.Errorf("create daily memory dir: %w", err)
	}

	date := time.Now().Format(time.DateOnly)
	filePath := filepath.Join(dm.dirPath, date+dailyMemoryFileExt)
	return os.WriteFile(filePath, []byte(content), 0o644)
}
