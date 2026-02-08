## Why

`pkg/skills` 中的日志传递方式与项目标准不一致。项目中绝大多数地方使用 `logr.FromContextOrDiscard(ctx)` 从 context 中获取 logger，但 `SkillLoader` 采用在构造函数中传入 logger 并存储在 struct 字段中的方式。这种模式无法支持日志链路追踪，且不符合 Go context 标准实践。

## What Changes

- **BREAKING**: 修改 `SkillLoader` 结构体，移除 `logger logr.Logger` 字段
- **BREAKING**: 修改 `NewSkillLoader` 和 `NewSkillLoaderWithDir` 构造函数签名，使用 `ctx context.Context` 参数替代 `logger logr.Logger` 参数
- 修改 `Load()` 方法签名，添加 `ctx context.Context` 参数
- `Load()` 方法内部使用 `logr.FromContextOrDiscard(ctx)` 获取 logger
- 修改调用方 `pkg/agents/genkit.go`，传递 context 而非 logger
- 修改测试文件 `pkg/skills/skill_loader_test.go` 和 `pkg/skills/skill_tool_test.go`，使用 `t.Context()`

**不影响的方法**: `Discover`, `Get`, `List`, `GetAll`, `SkillsDir`, `GetSkillMetadata`, `GetSkillContent` 这些方法不输出日志，无需添加 ctx 参数。

## Capabilities

### New Capabilities

无。这是内部重构，不引入新功能。

### Modified Capabilities

无。此变更仅修改实现细节，不改变 spec 级别的行为或要求。

## Impact

- **代码变更**:
  - `pkg/skills/skill_loader.go` - 主要修改
  - `pkg/agents/genkit.go` - 调用方修改
  - `pkg/skills/skill_loader_test.go` - 测试修改
  - `pkg/skills/skill_tool_test.go` - 测试修改

- **API 变更**: 构造函数签名和 `Load` 方法签名变更

- **无依赖变更**: 不涉及外部依赖或系统变更