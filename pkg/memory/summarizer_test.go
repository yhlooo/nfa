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
	m, err := Load(filepath.Join(t.TempDir(), "MEMORY.md"))
	require.NoError(t, err)

	called := false
	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		called = true
		return SummarizeOutput{Memory: "result"}, nil
	}

	s := &Summarizer{
		memory: m,
		run:    run,
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

	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		return SummarizeOutput{}, errors.New("model error")
	}

	s := &Summarizer{
		memory: m,
		run:    run,
	}

	s.summarize(context.Background(), []*ai.Message{ai.NewUserTextMessage("hello")})
	assert.Equal(t, "old content", m.Content())
}

func TestSummarizer_SuccessfulSummarization(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "MEMORY.md")
	require.NoError(t, os.WriteFile(filePath, []byte("old memory"), 0o644))

	m, err := Load(filePath)
	require.NoError(t, err)

	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		assert.Contains(t, in.ExistingMemory, "old memory")
		assert.NotEmpty(t, in.History)
		return SummarizeOutput{Memory: "# NFA Memory\n## 用户关注\n- 关注股票：TSLA"}, nil
	}

	s := &Summarizer{
		memory: m,
		run:    run,
	}
	s.summarize(context.Background(), []*ai.Message{ai.NewUserTextMessage("分析TSLA")})

	assert.Equal(t, "# NFA Memory\n## 用户关注\n- 关注股票：TSLA", m.Content())

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "# NFA Memory\n## 用户关注\n- 关注股票：TSLA", string(data))
}

func TestSummarizer_ContextCancelled(t *testing.T) {
	m, err := Load(filepath.Join(t.TempDir(), "MEMORY.md"))
	require.NoError(t, err)

	called := false
	run := func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
		called = true
		return SummarizeOutput{Memory: "result"}, nil
	}

	s := &Summarizer{
		memory: m,
		run:    run,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.summarize(ctx, []*ai.Message{ai.NewUserTextMessage("hello")})
	assert.False(t, called, "run should not be called when context is cancelled")
}
