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

// DailySummarizer 中期记忆异步总结器
type DailySummarizer struct {
	mu         sync.Mutex
	dailyMem   *DailyMemory
	cancelFunc context.CancelFunc
	run        func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error)
}

// NewDailySummarizer 创建中期记忆总结器
func NewDailySummarizer(g *genkit.Genkit, dailyMem *DailyMemory) *DailySummarizer {
	return &DailySummarizer{
		dailyMem: dailyMem,
		run:      DefineDailySummarizeFlow(g, "DailyMemorySummarize").Run,
	}
}

// Schedule 安排防抖总结。在调用方对话结束后调用。
// 等待 1 分钟，期间如果被 Cancel 则取消，否则开始异步总结。
func (s *DailySummarizer) Schedule(ctx context.Context, history []*ai.Message) {
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
func (s *DailySummarizer) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
}

// summarize 执行中期记忆总结
func (s *DailySummarizer) summarize(ctx context.Context, history []*ai.Message) {
	logger := logr.FromContextOrDiscard(ctx)

	select {
	case <-ctx.Done():
		return
	default:
	}

	existingMemory := ""
	// 加载当天已有记忆作为合并基础
	data, err := s.dailyMem.LoadRecent(1) // 只加载今天
	if err == nil {
		existingMemory = data
	}
	// 忽略加载错误，继续总结

	out, err := s.run(ctx, SummarizeInput{
		ExistingMemory: existingMemory,
		History:        history,
	})
	if err != nil {
		logger.Error(err, "daily memory summarization failed")
		return
	}

	result := strings.TrimSpace(out.Memory)
	if result == "" || result == existingMemory {
		return
	}

	if err := s.dailyMem.UpdateToday(result); err != nil {
		logger.Error(err, "failed to update daily memory file")
	}
}
