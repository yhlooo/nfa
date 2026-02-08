## Context

当前 `pkg/skills` 包中的 `SkillLoader` 采用传统的 logger 传递方式：在构造函数中接收 `logr.Logger` 参数并存储在结构体字段中，方法内部通过 `sl.logger.Info/Error` 记录日志。

然而，项目中其他所有模块都遵循标准 Go 实践，使用 `logr.NewContext(ctx, logger)` 和 `logr.FromContextOrDiscard(ctx)` 从 context 中传递和获取 logger。这种不一致性带来以下问题：

1. **无法支持链路追踪**：日志无法携带 request ID 等上下文信息
2. **不符合项目标准**：代码风格不统一，增加维护成本
3. **测试复杂度**：无法使用 `t.Context()` 获得更好的测试体验

约束条件：
- 仅限项目内部重构，无外部 API 兼容性要求
- 保持 `skill-system` spec 中定义的所有行为要求不变

## Goals / Non-Goals

**Goals:**
- 统一 `pkg/skills` 的日志传递方式与项目标准一致
- 使 `Load` 方法支持从 context 获取 logger
- 更新所有测试使用 `t.Context()`

**Non-Goals:**
- 不修改 `skill-system` spec 中的任何行为要求
- 不改变 API 的对外行为（仅内部实现细节）
- 不修改 `SkillLoader` 中不需要日志的方法签名

## Decisions

### 决策 1: 仅修改需要日志的方法

**选择**: 只为 `Load` 方法添加 `ctx context.Context` 参数

**理由**: `Discover`, `Get`, `List`, `GetAll`, `SkillsDir`, `GetSkillMetadata`, `GetSkillContent` 这些方法不输出日志，无需添加 ctx 参数。这样可以最小化 API 变更，保持接口简洁。

### 决策 2: 移除 struct 中的 logger 字段

**选择**: 从 `SkillLoader` 结构体中完全移除 `logger logr.Logger` 字段

**理由**: 如果保留该字段，开发者可能会继续使用它而不是 context。彻底移除可以强制使用新的方式，避免出现两套日志传递方式并存的情况。

### 决策 3: 构造函数接收 context.Context

**选择**: 构造函数签名改为 `NewSkillLoader(ctx context.Context, ...)` 和 `NewSkillLoaderWithDir(ctx context.Context, ...)`

**理由**: 虽然构造函数不立即需要 logger，但接收 context 为未来可能的初始化日志提供了灵活性，并且与项目其他构造函数（如 `NFAAgent.InitGenkit(ctx)`）保持一致的模式。

### 决策 4: 测试使用 t.Context()

**选择**: 所有测试调用 `NewSkillLoader(t.Context(), ...)` 和 `loader.Load(t.Context())`

**理由**: Go 1.24+ 提供的 `t.Context()` 会在测试失败时自动取消 context，有助于资源清理和日志关联。这是 Go 社区推荐的最佳实践。

## Risks / Trade-offs

**风险**: 如果项目中仍有未发现的直接使用 `NewSkillLoader(logger, ...)` 的代码，编译会失败

**缓解**: 项目不会被其他项目引用，仅需确保内部所有调用点都已更新。`grep` 搜索确认调用点只有 `pkg/agents/genkit.go` 和测试文件。

**权衡**: 每次调用 `Load(ctx)` 都会执行 `logr.FromContextOrDiscard(ctx)`，有轻微性能开销

**缓解**: `Load` 方法仅在初始化时调用一次（agent 启动时），不是高频调用，性能影响可忽略。

## Migration Plan

无需部署计划，这是纯代码重构。实施步骤：

1. 修改 `pkg/skills/skill_loader.go`
2. 修改 `pkg/agents/genkit.go`
3. 修改测试文件
4. 运行测试确保通过

## Open Questions

无