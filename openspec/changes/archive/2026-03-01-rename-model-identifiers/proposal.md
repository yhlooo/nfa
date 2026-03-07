# 重命名模型标识符

## Why

当前的模型标识符 `main` 和 `fast` 命名语义不够准确和正式：
- `main` 可能在某些上下文中引起混淆（如 main 函数、main 分支）
- `fast` 过于强调速度，但实际上该模型的核心特性是轻量级（低成本、低资源消耗）
- `primary` 和 `light` 能更准确地表达模型的定位和用途

## What Changes

**BREAKING**: 这是一个破坏性变更，将全面重命名模型相关标识符。

### 核心命名变更
- `main` → `primary` (主模型)
- `fast` → `light` (轻量模型)

### 具体变更范围

**代码层面**：
- Go 结构体字段和方法
  - `pkg/models/model_routing.go`: `Models.Main` → `Models.Primary`, `GetMain()` → `GetPrimary()`
  - `pkg/models/model_routing.go`: `Models.Fast` → `Models.Light`, `GetFast()` → `GetLight()`
- 命令行参数
  - `--fast-model` → `--light-model`
  - `--model` 保持不变（对应 primary）
- 交互式命令
  - `/model :main` → `/model :primary`
  - `/model :fast` → `/model :light`
- UI 常量和元数据键
  - `ModelTypeMain` → `ModelTypePrimary`
  - `ModelTypeFast` → `ModelTypeLight`
  - `v["main"]` → `v["primary"]`
  - `v["fast"]` → `v["light"]`

**配置层面**：
- `~/.nfa/nfa.json` 配置字段
  - `defaultModels.main` → `defaultModels.primary`
  - `defaultModels.fast` → `defaultModels.light`

**文档层面**：
- 更新所有相关文档中的模型命名引用
- 更新命令行示例和配置示例

## Capabilities

### Modified Capabilities
- `model-selection`: 交互式模型选择命令和显示文本中的标识符命名
  - 命令参数从 `:main`/:fast` 改为 `:primary`/:light`
  - 菜单标题从 "Select main/fast model" 改为 "Select primary/light model"
  - 成功消息从 "main/fast model set to" 改为 "primary/light model set to"

- `model-config`: 配置文件中的模型标识符字段命名
  - 配置字段从 `main`/`fast` 改为 `primary`/`light`

## Impact

### 受影响系统
- **代码库**: 7+ Go 源文件需要修改
- **配置文件**: 所有用户的 `~/.nfa/nfa.json` 需要手动更新
- **文档**: 5 个文档文件需要全面更新
- **用户习惯**: 交互命令和 CLI 参数的变化需要用户适应

### 用户迁移路径
由于不需要向后兼容，用户需要：
1. 更新配置文件中的字段名（`main` → `primary`, `fast` → `light`）
2. 适应新的命令行参数（`--fast-model` → `--light-model`）
3. 适应新的交互命令（`/model :main` → `/model :primary`, `/model :fast` → `/model :light`）

### 不受影响的部分
- `vision` 模型标识符保持不变
- 模型提供商配置不受影响
- 模型能力描述（reasoning, vision 等）不受影响
