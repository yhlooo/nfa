package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"
)

// Summarizer 异步记忆总结器
type Summarizer struct {
	mu         sync.Mutex
	memory     *Memory
	cancelFunc context.CancelFunc
	run        func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error)
}

// NewSummarizer 创建总结器
func NewSummarizer(g *genkit.Genkit, memory *Memory) *Summarizer {
	return &Summarizer{
		memory: memory,
		run:    DefineSummarizeFlow(g, "MemorySummarize").Run,
	}
}

// Schedule 安排防抖总结。在调用方对话结束后调用。
// 等待 1 分钟，期间如果被 Cancel 则取消，否则开始异步总结。
func (s *Summarizer) Schedule(ctx context.Context, history []*ai.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	ctx, cancel := context.WithCancel(logr.NewContext(context.Background(), logr.FromContextOrDiscard(ctx)))
	s.cancelFunc = cancel

	go func() {
		select {
		case <-time.After(1 * time.Minute):
		case <-ctx.Done():
			return
		}

		s.summarize(ctx, history)
	}()
}

// Cancel 取消正在进行的防抖等待或异步总结
func (s *Summarizer) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
}

// summarize 执行记忆总结
func (s *Summarizer) summarize(ctx context.Context, history []*ai.Message) {
	logger := logr.FromContextOrDiscard(ctx)

	select {
	case <-ctx.Done():
		return
	default:
	}

	existingMemory := s.memory.Content()

	out, err := s.run(ctx, SummarizeInput{
		ExistingMemory: existingMemory,
		History:        history,
	})
	if err != nil {
		logger.Error(err, "memory summarization failed")
		return
	}

	result := strings.TrimSpace(out.Memory)
	if result == "" || result == existingMemory {
		return
	}

	if err := s.memory.Update(result); err != nil {
		logger.Error(err, "failed to update memory file")
	}
}
