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
