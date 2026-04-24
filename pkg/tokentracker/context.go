package tokentracker

import (
	"context"

	"github.com/firebase/genkit/go/ai"
)

type trackerContextKey struct{}

// ContextWithTokenTracker 返回包含 Token 跟踪器的上下文
func ContextWithTokenTracker(ctx context.Context, t *TokenTracker) context.Context {
	return context.WithValue(ctx, trackerContextKey{}, t)
}

// TrackerFromContext 从上下文获取 Token 跟踪器
func TrackerFromContext(ctx context.Context) *TokenTracker {
	t, _ := ctx.Value(trackerContextKey{}).(*TokenTracker)
	return t
}

// ModelMiddlewareFromContext 从上下文获取 Token 跟踪器的模型中间件
func ModelMiddlewareFromContext(ctx context.Context, modelName string) ai.ModelMiddleware {
	t := TrackerFromContext(ctx)
	return t.ModelMiddleware(modelName)
}
