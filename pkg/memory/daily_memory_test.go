package memory

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDailyMemory_LoadRecent_NoFiles(t *testing.T) {
	dm := NewDailyMemory(t.TempDir())
	content, err := dm.LoadRecent(7)
	require.NoError(t, err)
	assert.Equal(t, "", content)
}

func TestDailyMemory_LoadRecent_WithFiles(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	// 创建今天和昨天的文件
	today := time.Now().Format(time.DateOnly)
	yesterday := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "memory"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "memory", today+".md"), []byte("# "+today+" memory\n- item today"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "memory", yesterday+".md"), []byte("# "+yesterday+" memory\n- item yesterday"), 0o644))

	content, err := dm.LoadRecent(7)
	require.NoError(t, err)

	assert.Contains(t, content, "item today")
	assert.Contains(t, content, "item yesterday")
	// 最新日期应在前
	assert.True(t, strings.Index(content, today) < strings.Index(content, yesterday),
		"today should appear before yesterday (newest first)")
}

func TestDailyMemory_LoadRecent_ExceedsAvailableDays(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	// 只创建今天的文件
	today := time.Now().Format(time.DateOnly)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "memory"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "memory", today+".md"), []byte("today's memory"), 0o644))

	content, err := dm.LoadRecent(30)
	require.NoError(t, err)
	assert.Equal(t, "today's memory", content)
}

func TestDailyMemory_UpdateToday_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	err := dm.UpdateToday("# Test\n- first entry")
	require.NoError(t, err)

	today := time.Now().Format(time.DateOnly)
	filePath := filepath.Join(dir, "memory", today+".md")
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "# Test\n- first entry", string(data))
}

func TestDailyMemory_UpdateToday_OverwritesFile(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	require.NoError(t, dm.UpdateToday("first content"))

	// 再次更新同一天
	require.NoError(t, dm.UpdateToday("second content"))

	content, err := dm.LoadRecent(1)
	require.NoError(t, err)
	assert.Equal(t, "second content", content)
}

func TestDailyMemory_UpdateToday_AutoCreatesDir(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	// DirPath 不存在，无人工创建
	err := dm.UpdateToday("auto created")
	require.NoError(t, err)

	today := time.Now().Format(time.DateOnly)
	filePath := filepath.Join(dir, "memory", today+".md")
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "auto created", string(content))
}

func TestDailyMemory_ConcurrentReadWrite(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = dm.LoadRecent(7)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = dm.UpdateToday("concurrent update")
	}()
	wg.Wait()
	// 无 panic 即为通过
}

func TestDailyMemory_LoadRecent_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	today := time.Now().Format(time.DateOnly)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "memory"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "memory", today+".md"), []byte("  \n  "), 0o644))

	content, err := dm.LoadRecent(1)
	require.NoError(t, err)
	assert.Equal(t, "", content)
}
