package flows

import (
	"context"

	"github.com/firebase/genkit/go/ai"
)

type modelNameContextKey struct{}

// ContextWithModelName 返回携带指定模型名的上下文
func ContextWithModelName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, modelNameContextKey{}, name)
}

// ModelNameFromContext 从上下文获取模型名
func ModelNameFromContext(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(modelNameContextKey{}).(string)
	return name, ok
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
