package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Store 记忆存储
type Store interface {
	// Read 读取记忆内容
	Read() (string, error)
	// ReadActiveContent 读活跃部分
	ReadActiveContent() (string, error)
	// Update 更新记忆内容
	Update(content string) error
}

// FileStore 基于文件的记忆存储
type FileStore struct {
	path string

	lock         sync.Mutex
	cacheContent string
	cacheOk      bool
}

// NewFileStore 基于文件的记忆存储
func NewFileStore(path string) *FileStore {
	return &FileStore{
		path: path,
	}
}

// Read 读取记忆内容
func (s *FileStore) Read() (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// 返回缓存
	if s.cacheOk {
		return s.cacheContent, nil
	}

	// 读取文件
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.cacheOk = true
			return "", nil
		}
		return "", err
	}

	s.cacheContent = string(data)
	s.cacheOk = true

	return s.cacheContent, nil
}

// ReadActiveContent 读活跃部分
func (s *FileStore) ReadActiveContent() (string, error) {
	return s.Read()
}

// Update 更新记忆内容
func (s *FileStore) Update(content string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cacheContent = content
	return os.WriteFile(s.path, []byte(content), 0o644)
}

// DailyFilesStore 每天一个文件存储
type DailyFilesStore struct {
	dir      string
	readDays int

	lock          sync.Mutex
	activeContent string
	activeDate    string
}

// NewDailyStore 创建 *DailyFilesStore
func NewDailyStore(dir string, readDays int) *DailyFilesStore {
	if readDays <= 0 {
		readDays = 7
	}
	return &DailyFilesStore{
		dir:      dir,
		readDays: readDays,
	}
}

// Read 读取最近一段时间记忆内容
func (s *DailyFilesStore) Read() (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	today := time.Now()

	ret := &strings.Builder{}
	for i := 0; i < s.readDays; i++ {
		date := today.AddDate(0, 0, -i).Format(time.DateOnly)
		path := filepath.Join(s.dir, date+".md")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("read daily memory %s error: %w", date, err)
		}
		content := strings.TrimSpace(string(data))
		if content != "" {
			ret.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", date, content))
		}
	}

	return ret.String(), nil
}

// ReadActiveContent 读活跃部分
func (s *DailyFilesStore) ReadActiveContent() (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	today := time.Now().Format(time.DateOnly)
	if s.activeDate == today {
		return s.activeContent, nil
	}

	// 读取文件
	data, err := os.ReadFile(filepath.Join(s.dir, today+".md"))
	if err != nil {
		if os.IsNotExist(err) {
			s.activeDate = today
			return "", nil
		}
		return "", err
	}

	s.activeContent = string(data)
	s.activeDate = today

	return s.activeContent, nil
}

// Update 写入当天记忆内容
func (s *DailyFilesStore) Update(content string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	today := time.Now().Format(time.DateOnly)

	s.activeContent = content
	s.activeDate = today

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create daily memory dir: %w", err)
	}
	filePath := filepath.Join(s.dir, today+".md")
	return os.WriteFile(filePath, []byte(content), 0o644)
}
