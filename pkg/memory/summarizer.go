package memory

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"
)

// Summarizer 异步记忆总结器，统一管理长期记忆和中期记忆的总结
type Summarizer struct {
	lock       sync.Mutex
	cancelFunc context.CancelFunc

	summarizeLongTermMemoryFlow SummarizeFlow
	longTermMemory              Store

	summarizeDailyMemoryFlow SummarizeFlow
	dailyMemory              Store
}

// NewSummarizer 创建记忆总结器
func NewSummarizer(g *genkit.Genkit, dataRoot string) *Summarizer {
	return &Summarizer{
		summarizeLongTermMemoryFlow: DefineSummarizeFlow(g, "SummarizeLongTermMemory", SummarizeOptions{
			SystemPrompt: summarizeLongTermMemorySystemPrompt,
		}),
		longTermMemory: NewFileStore(filepath.Join(dataRoot, "MEMORY.md")),
		summarizeDailyMemoryFlow: DefineSummarizeFlow(g, "SummarizeDailyMemory", SummarizeOptions{
			SystemPrompt: summarizeDailyMemorySystemPrompt,
		}),
		dailyMemory: NewDailyStore(filepath.Join(dataRoot, "memory"), 7),
	}
}

// Schedule 安排防抖总结。在调用方对话结束后调用。
// 等待 1 分钟，期间如果被 Cancel 则取消，否则开始异步总结。
func (s *Summarizer) Schedule(ctx context.Context, history []*ai.Message) {
	s.lock.Lock()
	defer s.lock.Unlock()

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

		// 总结长期记忆
		s.summarize(ctx, history, s.summarizeLongTermMemoryFlow, s.longTermMemory)
		// 总结每日记忆
		s.summarize(ctx, history, s.summarizeDailyMemoryFlow, s.dailyMemory)
	}()
}

// Cancel 取消正在进行的防抖等待或异步总结
func (s *Summarizer) Cancel() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
}

// LongTermMemory 返回长期记忆
func (s *Summarizer) LongTermMemory() Store {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.longTermMemory
}

// DailyMemory 返回每日记忆
func (s *Summarizer) DailyMemory() Store {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.dailyMemory
}

// summarize 执行单个流程的记忆总结
func (s *Summarizer) summarize(ctx context.Context, history []*ai.Message, flow SummarizeFlow, mem Store) {
	logger := logr.FromContextOrDiscard(ctx)

	select {
	case <-ctx.Done():
		return
	default:
	}

	activeMemory, err := mem.ReadActiveContent()
	if err != nil {
		logger.Error(err, "read active memory for summarization error")
		return
	}

	out, err := flow.Run(ctx, SummarizeInput{
		CurrentMemory: activeMemory,
		History:       history,
	})
	if err != nil {
		logger.Error(err, "memory summarization error")
		return
	}

	newMem := strings.TrimSpace(out.Memory) + "\n"
	if newMem == "" || newMem == activeMemory {
		return
	}

	if err := mem.Update(newMem); err != nil {
		logger.Error(err, "save new memory error")
	}
}
