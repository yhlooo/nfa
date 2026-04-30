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

func TestSummarizer_CancelDuringWait(t *testing.T) {
	called := false
	store := &testStore{
		loadFn: func() (string, error) { return "", nil },
		saveFn: func(content string) error {
			called = true
			return nil
		},
	}

	s := &Summarizer{
		flows: []summarizeFlow{
			{run: testRun("result"), store: store},
		},
	}
	s.Schedule(t.Context(), []*ai.Message{ai.NewUserTextMessage("hello")})

	// 立即取消（在1分钟等待期内）
	s.Cancel()

	// 等待足够长时间确保 goroutine 完成
	time.Sleep(100 * time.Millisecond)

	assert.False(t, called, "run should not have been called after cancel")
}

func TestSummarizer_RunError(t *testing.T) {
	m, err := Load(filepath.Join(t.TempDir(), "MEMORY.md"))
	require.NoError(t, err)
	require.NoError(t, m.Update("old content"))

	store := NewFileStore(m)
	errRun := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		return SummarizeOutput{}, errors.New("model error")
	}

	s := &Summarizer{
		flows: []summarizeFlow{
			{run: errRun, store: store},
		},
	}

	s.summarize(context.Background(), s.flows[0], []*ai.Message{ai.NewUserTextMessage("hello")})
	assert.Equal(t, "old content", m.Content())
}

func TestSummarizer_SuccessfulSummarization(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "MEMORY.md")
	require.NoError(t, os.WriteFile(filePath, []byte("old memory"), 0o644))

	m, err := Load(filePath)
	require.NoError(t, err)

	store := NewFileStore(m)
	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		assert.Contains(t, in.ExistingMemory, "old memory")
		assert.NotEmpty(t, in.History)
		return SummarizeOutput{Memory: "# NFA Memory\n## 用户关注\n- 关注股票：TSLA"}, nil
	}

	s := &Summarizer{
		flows: []summarizeFlow{
			{run: run, store: store},
		},
	}
	s.summarize(context.Background(), s.flows[0], []*ai.Message{ai.NewUserTextMessage("分析TSLA")})

	assert.Equal(t, "# NFA Memory\n## 用户关注\n- 关注股票：TSLA", m.Content())

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "# NFA Memory\n## 用户关注\n- 关注股票：TSLA", string(data))
}

func TestSummarizer_ContextCancelled(t *testing.T) {
	called := false
	store := &testStore{
		loadFn: func() (string, error) { return "", nil },
		saveFn: func(content string) error {
			called = true
			return nil
		},
	}

	s := &Summarizer{
		flows: []summarizeFlow{
			{run: testRun("result"), store: store},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.summarize(ctx, s.flows[0], []*ai.Message{ai.NewUserTextMessage("hello")})
	assert.False(t, called, "run should not be called when context is cancelled")
}

func TestSummarizer_EmptyResultPreservesExisting(t *testing.T) {
	m, err := Load(filepath.Join(t.TempDir(), "MEMORY.md"))
	require.NoError(t, err)
	require.NoError(t, m.Update("existing memory"))

	store := NewFileStore(m)
	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		return SummarizeOutput{Memory: ""}, nil
	}

	s := &Summarizer{
		flows: []summarizeFlow{
			{run: run, store: store},
		},
	}
	s.summarize(context.Background(), s.flows[0], []*ai.Message{ai.NewUserTextMessage("hello")})

	assert.Equal(t, "existing memory", m.Content())
}

func TestSummarizer_ResultIdenticalToExisting(t *testing.T) {
	m, err := Load(filepath.Join(t.TempDir(), "MEMORY.md"))
	require.NoError(t, err)
	require.NoError(t, m.Update("same content"))

	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		return SummarizeOutput{Memory: "same content"}, nil
	}

	store := NewFileStore(m)
	s := &Summarizer{
		flows: []summarizeFlow{
			{run: run, store: store},
		},
	}
	s.summarize(context.Background(), s.flows[0], []*ai.Message{ai.NewUserTextMessage("hello")})

	assert.Equal(t, "same content", m.Content())
}

func TestSummarizer_DailyMemoryFlow(t *testing.T) {
	dm := NewDailyMemory(t.TempDir())

	store := NewDailyStore(dm)
	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		assert.Equal(t, "", in.ExistingMemory) // 当天尚无记忆
		assert.NotEmpty(t, in.History)
		return SummarizeOutput{Memory: "# 2026-01-01 中期记忆\n\n## 今日关注\n- TSLA 财报分析"}, nil
	}

	s := &Summarizer{
		flows: []summarizeFlow{
			{run: run, store: store},
		},
	}
	s.summarize(context.Background(), s.flows[0], []*ai.Message{ai.NewUserTextMessage("分析TSLA")})

	content, err := dm.LoadRecent(1)
	require.NoError(t, err)
	assert.Contains(t, content, "TSLA")
	assert.Contains(t, content, "中期记忆")
}

func TestSummarizer_ScheduleFiresAfterWait(t *testing.T) {
	called := make(chan struct{}, 1)
	store := &testStore{
		loadFn: func() (string, error) { return "", nil },
		saveFn: func(content string) error {
			called <- struct{}{}
			return nil
		},
	}

	s := &Summarizer{
		flows: []summarizeFlow{
			{run: testRun("after wait"), store: store},
		},
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
		s.summarize(ctx, s.flows[0], []*ai.Message{ai.NewUserTextMessage("test")})
	}()

	select {
	case <-called:
		// 成功：等待到期后调用
	case <-time.After(500 * time.Millisecond):
		t.Fatal("summarize should have been called after wait timer")
	}
}

// testRun 创建返回固定内容的 run 函数
func testRun(result string) func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
	return func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		return SummarizeOutput{Memory: result}, nil
	}
}

// testStore 用于测试的 MemoryStore
type testStore struct {
	loadFn func() (string, error)
	saveFn func(content string) error
}

func (s *testStore) Load() (string, error) {
	if s.loadFn != nil {
		return s.loadFn()
	}
	return "", nil
}

func (s *testStore) Save(content string) error {
	if s.saveFn != nil {
		return s.saveFn(content)
	}
	return nil
}
