# 重命名模型标识符 - 设计文档

## Context

### 当前状态
代码库中使用 `main` 和 `fast` 作为模型类型标识符，分布在以下层次：
- **数据结构层**: `Models` 结构体的字段和方法
- **配置层**: JSON 配置文件的 `defaultModels` 对象
- **命令行层**: CLI flags 和 TUI 交互命令
- **运行时层**: Agent 元数据传递和模型路由逻辑

### 约束条件
- 用户明确要求**不需要向后兼容**，允许直接破坏性变更
- `--model` 参数保持不变，仍映射到 primary 模型
- `vision` 模型标识符不在此次变更范围内
- 所有文档必须同步更新

### 利益相关者
- 现有用户：需要手动迁移配置文件
- 新用户：获得更清晰的命名体验
- 开发者：代码语义更明确

## Goals / Non-Goals

**Goals:**
- 全面重命名所有模型类型标识符（main → primary, fast → light）
- 更新所有相关代码、配置、文档
- 保持功能逻辑不变，仅修改命名
- 提供清晰的错误提示和用户反馈

**Non-Goals:**
- 不实现新的模型功能或特性
- 不改变模型选择或路由的逻辑
- 不提供自动配置迁移工具
- 不保持向后兼容性

## Decisions

### 1. Go 结构体和方法命名

**决策**: 直接重命名字段和方法，不使用别名或包装器

**理由**:
- 保持代码简洁一致，避免冗余的兼容层
- 用户明确不需要向后兼容
- 减少维护成本和代码复杂度

**实现**:
```go
// Before
type Models struct {
    Main   string `json:"main"`
    Fast   string `json:"fast"`
    Vision string `json:"vision"`
}
func (m Models) GetMain() string
func (m Models) GetFast() string

// After
type Models struct {
    Primary string `json:"primary"`
    Light   string `json:"light"`
    Vision  string `json:"vision"`
}
func (m Models) GetPrimary() string
func (m Models) GetLight() string
```

### 2. CLI 参数命名

**决策**: `--model` 保持不变，`--fast-model` 改为 `--light-model`

**理由**:
- `--model` 已广泛使用，改动影响面大
- `--model` 映射到 primary 模型在语义上仍然合理
- `--light-model` 明确表达轻量模型的概念

**替代方案考虑**:
- 选项 A: 同时改名为 `--primary-model` 和 `--light-model`
  - 缺点: 参数名变长，命令行输入更繁琐
- 选项 B: 保持两个参数都不变，仅改变内部实现
  - 缺点: 外部接口和内部概念不一致，造成混淆

### 3. 交互式命令前缀

**决策**: 将 `:main` 和 `:fast` 改为 `:primary` 和 `:light`

**理由**:
- 与新的标识符命名保持一致
- 长度增加但语义更清晰（:primary 比 :main 长 4 字符，:light 比 :fast 长 1 字符）

**影响**:
```go
// Before
case "/model", "/model :main"
case "/model :fast"

// After
case "/model", "/model :primary"
case "/model :light"
```

### 4. 配置文件字段命名

**决策**: 直接修改 JSON 字段名，不提供迁移工具

**理由**:
- 配置文件通常由高级用户手动编辑
- 字段名简单直接的查找替换即可
- 避免过度工程化

**用户影响**:
用户需要手动编辑 `~/.nfa/nfa.json`:
```json
// Before
{
  "defaultModels": {
    "main": "ollama/llama3.2",
    "fast": "ollama/mistral"
  }
}

// After
{
  "defaultModels": {
    "primary": "ollama/llama3.2",
    "light": "ollama/mistral"
  }
}
```

### 5. 元数据键命名

**决策**: 将 Agent 元数据中的键名从 "main"/"fast" 改为 "primary"/"light"

**实现**:
```go
// Before
v["main"]
v["fast"]

// After
v["primary"]
v["light"]
```

### 6. UI 显示文本

**决策**: 所有用户可见的显示文本同步更新为新的命名

**范围**:
- 模型选择菜单标题
- 成功提示消息
- 错误提示消息
- 模型类型显示（如 "ollama/llama3.2 (primary)"）

## Risks / Trade-offs

### 风险 1: 配置文件解析失败导致启动错误
**描述**: 用户更新代码但未更新配置文件时，JSON 解析可能忽略未知字段或使用零值

**缓解措施**:
- 在配置加载时添加验证，如果 primary/light 字段缺失且存在 main/fast 字段，提供清晰的错误提示
- 错误消息应包含迁移指南："配置文件已更新，请将 'main' 改为 'primary', 'fast' 改为 'light'"

### 风险 2: 文档遗漏导致不一致
**描述**: 某些文档中的示例或说明可能未及时更新，造成用户困惑

**缓解措施**:
- 使用全局搜索确保所有文档中的 main/fast 引用都已更新
- 在 PR review 时重点检查文档同步性

### 风险 3: 用户脚本和别名失效
**描述**: 用户可能有使用旧参数的 shell 别名或脚本

**缓解措施**:
- 在 CHANGELOG 或发布说明中明确说明参数变化
- 提供迁移示例（如：`alias nfa-quick='nfa --light-model ollama/mistral'`）

### 权衡: 便利性 vs 一致性
**决策**: 选择命名一致性和语义清晰度，牺牲一定的命令行输入便利性

**影响**:
- 正面: 代码和文档更易理解，新概念学习曲线降低
- 负面: 交互命令输入稍微变长（但仅在手动输入时有感知）

## Migration Plan

### 阶段 1: 代码变更
1. 修改 Go 结构体和方法（`pkg/models/model_routing.go`）
2. 修改 CLI 参数（`pkg/commands/root.go`）
3. 修改 UI 常量和命令处理（`pkg/ui/chat/ui.go`）
4. 修改 Agent 元数据键（`pkg/agents/meta.go`）
5. 运行 `go fmt ./...` 和 `go vet ./...` 确保代码质量
6. 运行 `go test ./...` 确保测试通过

### 阶段 2: 文档更新
1. 更新 `docs/reference/config.md`
2. 更新 `docs/guides/command-line.md`
3. 更新 `docs/guides/model-config.md`
4. 更新 `openspec/specs/model-selection/spec.md`
5. 更新 `CLAUDE.md`

### 阶段 3: 验证和测试
1. 手动测试交互式模型选择命令
2. 手动测试 CLI 参数
3. 验证配置文件加载和保存
4. 验证所有文档示例的正确性

### 回滚策略
如果发现问题，可以通过 git revert 快速回滚到变更前的版本。由于不需要兼容旧配置，回滚后用户需将配置文件改回 `main`/`fast`。

## Open Questions

**Q1: 是否需要在配置加载时提供友好的迁移错误提示？**

A: 是的，建议添加。当检测到配置中存在 `main`/`fast` 字段但缺失 `primary`/`light` 字段时，显示错误提示："配置文件需要更新：将 'main' 改为 'primary', 'fast' 改为 'light'"

**Q2: 是否需要更新模型列表命令的输出格式？**

A: 是的，如果模型列表显示中包含模型类型标识（如 "(main)", "(fast)"），需要同步更新为 "(primary)", "(light)"

**Q3: 日志文件中的模型类型标识是否需要更新？**

A: 建议更新。日志作为调试和问题排查的重要依据，应使用一致的命名
