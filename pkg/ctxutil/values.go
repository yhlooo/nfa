package ctxutil

import (
	"context"

	"github.com/firebase/genkit/go/ai"

	"github.com/yhlooo/nfa/pkg/agents/models"
)

type modelsContextKey struct{}

// ContextWithModels 返回携带指定模型配置的上下文
func ContextWithModels(ctx context.Context, m models.Models) context.Context {
	return context.WithValue(ctx, modelsContextKey{}, m)
}

// ModelsFromContext 从上下文获取模型名
func ModelsFromContext(ctx context.Context) (models.Models, bool) {
	m, ok := ctx.Value(modelsContextKey{}).(models.Models)
	return m, ok
}

type handleStreamFnContextKey struct{}

// ContextWithHandleStreamFn 返回携带处理流函数的上下文
func ContextWithHandleStreamFn(ctx context.Context, fn ai.ModelStreamCallback) context.Context {
	return context.WithValue(ctx, handleStreamFnContextKey{}, fn)
}

// HandleStreamFnFromContext 从上下文获取处理流函数
func HandleStreamFnFromContext(ctx context.Context) ai.ModelStreamCallback {
	fn, ok := ctx.Value(handleStreamFnContextKey{}).(ai.ModelStreamCallback)
	if !ok {
		return nil
	}
	return fn
}

type modelUsageKey struct{}

// ContextWithModelUsage 返回携带模型用量信息的上下文
func ContextWithModelUsage(ctx context.Context, usage ai.GenerationUsage) context.Context {
	return context.WithValue(ctx, modelUsageKey{}, &usage)
}

// AddModelUsageToContext 追加模型用量到上下文
func AddModelUsageToContext(ctx context.Context, usage *ai.GenerationUsage) bool {
	if usage == nil {
		return false
	}

	cur, ok := ctx.Value(modelUsageKey{}).(*ai.GenerationUsage)
	if !ok || cur == nil {
		return false
	}

	if usage.Custom != nil {
		if cur.Custom == nil {
			cur.Custom = make(map[string]float64)
		}
		for k, v := range usage.Custom {
			cur.Custom[k] += v
		}
	}

	cur.CachedContentTokens += usage.CachedContentTokens
	cur.InputAudioFiles += usage.InputAudioFiles
	cur.InputCharacters += usage.InputCharacters
	cur.InputImages += usage.InputImages
	cur.InputTokens += usage.InputTokens
	cur.InputVideos += usage.InputVideos
	cur.OutputAudioFiles += usage.OutputAudioFiles
	cur.OutputCharacters += usage.OutputCharacters
	cur.OutputImages += usage.OutputImages
	cur.OutputTokens += usage.OutputTokens
	cur.OutputVideos += usage.OutputVideos
	cur.ThoughtsTokens += usage.ThoughtsTokens
	cur.TotalTokens += usage.TotalTokens

	return true
}

// GetModelUsageFromContext 从上下文获取当前模型用量
func GetModelUsageFromContext(ctx context.Context) ai.GenerationUsage {
	usage, ok := ctx.Value(modelUsageKey{}).(*ai.GenerationUsage)
	if !ok {
		return ai.GenerationUsage{}
	}
	return *usage
}
