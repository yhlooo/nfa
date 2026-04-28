package memory

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_FileNotExists(t *testing.T) {
	m, err := Load(filepath.Join(t.TempDir(), "nonexistent.md"))
	require.NoError(t, err)
	assert.Equal(t, "", m.Content())
}

func TestLoad_FileExists(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "MEMORY.md")
	err := os.WriteFile(filePath, []byte("# Test Memory\n- item 1"), 0o644)
	require.NoError(t, err)

	m, err := Load(filePath)
	require.NoError(t, err)
	assert.Equal(t, "# Test Memory\n- item 1", m.Content())
}

func TestUpdate_SavesAndUpdatesContent(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "MEMORY.md")

	// 创建初始文件
	err := os.WriteFile(filePath, []byte("old content"), 0o644)
	require.NoError(t, err)

	m, err := Load(filePath)
	require.NoError(t, err)
	assert.Equal(t, "old content", m.Content())

	// 更新
	err = m.Update("new content")
	require.NoError(t, err)

	// 验证内存中的内容已更新
	assert.Equal(t, "new content", m.Content())

	// 验证文件已更新
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "new content", string(data))
}

func TestMemory_ConcurrentReadWrite(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "MEMORY.md")

	m, err := Load(filePath)
	require.NoError(t, err)

	err = m.Update("initial")
	require.NoError(t, err)

	var wg sync.WaitGroup
	// 并发读取
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.Content()
		}()
	}
	// 并发写入
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = m.Update("updated")
	}()
	wg.Wait()

	// 最终内容应该是完整的字符串，不是部分写入
	content := m.Content()
	assert.True(t, content == "initial" || content == "updated",
		"content should be either old or new value, got: %q", content)
}
