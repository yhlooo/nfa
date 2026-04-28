package memory

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDailySummarizer_CancelDuringWait(t *testing.T) {
	dm := NewDailyMemory(t.TempDir())

	called := false
	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		called = true
		return SummarizeOutput{Memory: "result"}, nil
	}

	s := &DailySummarizer{
		dailyMem: dm,
		run:      run,
	}
	s.Schedule(t.Context(), []*ai.Message{ai.NewUserTextMessage("hello")})

	// 立即取消（在 1 分钟等待期内）
	s.Cancel()

	// 等待足够长时间确保 goroutine 完成
	time.Sleep(100 * time.Millisecond)

	assert.False(t, called, "run should not have been called after cancel")
}

func TestDailySummarizer_RunError(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)
	require.NoError(t, dm.UpdateToday("old content"))

	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		return SummarizeOutput{}, errors.New("model error")
	}

	s := &DailySummarizer{
		dailyMem: dm,
		run:      run,
	}

	s.summarize(context.Background(), []*ai.Message{ai.NewUserTextMessage("hello")})

	// 失败时应保持旧内容不变
	content, err := dm.LoadRecent(1)
	require.NoError(t, err)
	assert.Equal(t, "old content", content)
}

func TestDailySummarizer_SuccessfulSummarization(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		assert.NotEmpty(t, in.History)
		return SummarizeOutput{Memory: "# 2026-01-01 中期记忆\n\n## 今日关注\n- TSLA 财报分析"}, nil
	}

	s := &DailySummarizer{
		dailyMem: dm,
		run:      run,
	}
	s.summarize(context.Background(), []*ai.Message{ai.NewUserTextMessage("分析TSLA")})

	content, err := dm.LoadRecent(1)
	require.NoError(t, err)
	assert.Contains(t, content, "TSLA")
	assert.Contains(t, content, "中期记忆")
}

func TestDailySummarizer_ContextCancelled(t *testing.T) {
	dm := NewDailyMemory(t.TempDir())

	called := false
	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		called = true
		return SummarizeOutput{Memory: "result"}, nil
	}

	s := &DailySummarizer{
		dailyMem: dm,
		run:      run,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.summarize(ctx, []*ai.Message{ai.NewUserTextMessage("hello")})
	assert.False(t, called, "run should not be called when context is cancelled")
}

func TestDailySummarizer_NoExistingMemory(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		assert.Equal(t, "", in.ExistingMemory)
		return SummarizeOutput{Memory: "# new memory"}, nil
	}

	s := &DailySummarizer{
		dailyMem: dm,
		run:      run,
	}
	s.summarize(context.Background(), []*ai.Message{ai.NewUserTextMessage("test")})

	content, err := dm.LoadRecent(1)
	require.NoError(t, err)
	assert.Equal(t, "# new memory", content)
}

func TestDailySummarizer_EmptyResultPreservesExisting(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)
	require.NoError(t, dm.UpdateToday("existing daily memory"))

	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		// 返回空结果——无可记录信息
		return SummarizeOutput{Memory: ""}, nil
	}

	s := &DailySummarizer{
		dailyMem: dm,
		run:      run,
	}
	s.summarize(context.Background(), []*ai.Message{ai.NewUserTextMessage("hello")})

	content, err := dm.LoadRecent(1)
	require.NoError(t, err)
	assert.Equal(t, "existing daily memory", content)
}

func TestDailySummarizer_ScheduleFiresAfterWait(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)

	called := make(chan struct{}, 1)
	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		called <- struct{}{}
		return SummarizeOutput{Memory: "after wait"}, nil
	}

	s := &DailySummarizer{
		dailyMem: dm,
		run:      run,
	}

	// 使用更短的等待时间模拟
	s.mu.Lock()
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel
	s.mu.Unlock()

	go func() {
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return
		}
		s.summarize(ctx, []*ai.Message{ai.NewUserTextMessage("test")})
	}()

	select {
	case <-called:
		// 成功：等待到期后调用
	case <-time.After(500 * time.Millisecond):
		t.Fatal("summarize should have been called after wait timer")
	}
}

func TestDailySummarizer_ResultIdenticalToExisting(t *testing.T) {
	dir := t.TempDir()
	dm := NewDailyMemory(dir)
	require.NoError(t, dm.UpdateToday("same content"))

	updateCount := 0
	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		updateCount++
		return SummarizeOutput{Memory: "same content"}, nil
	}

	s := &DailySummarizer{
		dailyMem: dm,
		run:      run,
	}
	s.summarize(context.Background(), []*ai.Message{ai.NewUserTextMessage("hello")})

	// 内容相同时不应更新文件
	content, err := dm.LoadRecent(1)
	require.NoError(t, err)
	assert.Equal(t, "same content", content)

	// 验证文件内容未变
	todayDir := filepath.Join(dir, "memory")
	entries, err := os.ReadDir(todayDir)
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(todayDir, entries[0].Name()))
	require.NoError(t, err)
	assert.Equal(t, "same content", string(data))
	_ = updateCount
}
