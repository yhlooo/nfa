## Context

**当前状态**
NFA 系统通过 ACP (Agent Client Protocol) 实现 UI 和 Agent 的通信。现有的模型配置流程是：
1. 启动时从 `~/.nfa/nfa.json` 加载配置
2. 通过 `--model` 和 `--fast-model` 命令行参数覆盖
3. Agent 在 `Initialize` 时通过 Meta 返回 `availableModels` 和 `defaultModel`
4. UI 在 `newPrompt` 时通过 `PromptRequest.Meta.modelName` 覆盖主模型
5. Agent 的 `Prompt` 方法只支持从 Meta 读取主模型，fast 和 vision 模型使用配置文件的默认值

**架构约束**
- UI 层使用 Bubble Tea (elm-inspired architecture) 实现交互式终端界面
- Agent 和 UI 通过 ACP 协议通信，使用 `Meta` 字段传递元数据
- 模型提供商通过 Genkit 框架注册，返回模型名称字符串列表
- 配置文件使用 JSON 格式，由 `configs.LoadConfig` 加载（无保存功能）

**利益相关者**
- 终端用户：需要方便快捷地切换模型
- 配置管理员：需要配置持久化到文件

## Goals / Non-Goals

**Goals:**
- 运行时动态切换 main/fast/vision 三种模型类型
- 提供交互式选择菜单和直接命令两种方式
- 选择后立即保存配置到文件
- Agent 正确使用 UI 传递的模型配置
- 显示模型描述信息辅助用户选择

**Non-Goals:**
- 不修改配置文件格式（保持向后兼容）
- 不新增外部依赖
- 不改变现有 Agent 的模型路由逻辑（仅增加 Meta 读取）
- 不实现模型验证（不检查模型是否真正可用）
- 不支持批量切换或预设配置

## Decisions

### 1. UI 视图状态管理：模态视图替换

**决策**: ChatUI 使用 `viewState` 字段在 `input` 和 `model_select` 两个状态间切换，选择菜单出现时输入框完全隐藏。

**理由**:
- **简单直观**: 符合终端用户对"弹出菜单"的预期
- **状态隔离**: 选择菜单和输入框的键盘事件不会冲突（ESC/Enter 语义清晰）
- **易于实现**: 不需要管理复杂的焦点切换逻辑

**替代方案**:
- *焦点切换*: 输入框和选择菜单同时显示，通过 Tab 切换焦点
  - ❌ 复杂度高，容易产生焦点混乱
  - ❌ 选择菜单需要边框区分，视觉不简洁

### 2. Meta 扩展：独立字段方式

**决策**: 新增 `MetaKeyFastModel` 和 `MetaKeyVisionModel` 两个独立 Meta 键，与现有 `MetaKeyModelName` 并行。

**理由**:
- **向后兼容**: 现有的 `modelName` 键保持不变，不影响已有代码
- **类型安全**: 每个模型类型使用独立的字符串键，易于序列化和反序列化
- **简单直接**: Agent 端只需增加两行读取代码，无需解析复杂的 JSON 结构

**替代方案**:
- *JSON 对象序列化*: 使用 `models: {"main": "...", "fast": "..."}` 格式
  - ❌ 需要序列化/反序列化整个 Models 结构
  - ❌ 现有 `modelName` 键需要废弃或同步维护

### 3. 模型描述传递：JSON 映射方式

**决策**: Agent 在 `Initialize` 时返回 `MetaKeyModelDescriptions` (类型 `map[string]string`)，键为 `provider/name` 格式。

**理由**:
- **松耦合**: 描述信息独立于模型注册流程，不影响 Genkit 的模型定义
- **可扩展**: 未来可以添加更多元数据（如价格、上下文长度等）
- **序列化友好**: ACP 的 Meta 天然支持 map 类型

**替代方案**:
- *返回结构化数组*: `[]ModelInfo{Name, Description}`
  - ❌ 需要新增结构体定义
  - ❌ UI 端需要遍历数组查找描述，性能较差

### 4. 配置保存时机：立即保存

**决策**: 每次用户选择或设置模型后立即调用 `configs.SaveConfig` 保存到文件。

**理由**:
- **数据安全**: 避免用户关闭应用后丢失配置
- **用户预期**: 符合"设置立即生效"的直觉

**风险缓解**:
- 保存失败不影响 UI 继续使用，显示错误提示但不阻塞流程
- 使用原子写入（先写临时文件再重命名）避免损坏配置文件

### 5. 模型描述截断：80 字符限制

**决策**: ModelSelector 在渲染时将描述截断至 80 字符，超出部分使用 "..." 替代。

**理由**:
- **可读性**: 终端宽度通常为 80-120 字符，长描述会影响布局
- **一致性**: 统一的截断规则使界面更整洁

**替代方案**:
- *动态换行*: 根据终端宽度自动换行
  - ❌ 实现复杂， Bubble Tea 的窗口大小变化需要处理
  - ❌ 多行描述会影响选择项的对齐

### 6. ChatUI 配置管理：持有完整 Config

**决策**: ChatUI 新增 `cfg` 和 `cfgPath` 字段，持有完整的配置对象和路径。

**理由**:
- **避免重复加载**: 不需要在每次保存时重新从 context 获取配置
- **类型安全**: 直接修改 `cfg.DefaultModels` 而不是操作 map

**替代方案**:
- *每次保存时从 context 获取配置*
  - ❌ 需要在 context 中传递可变的配置对象，不符合 Go 的最佳实践

## Architecture

### 组件交互图

```
┌─────────────────────────────────────────────────────────────────┐
│                         ChatUI                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────┐     ┌──────────────┐                          │
│  │ InputBox     │────▶│ Command      │                          │
│  │              │     │ Parser       │                          │
│  └──────────────┘     └──────┬───────┘                          │
│                              │                                   │
│                              ▼                                   │
│                       ┌──────────────┐                          │
│                       │ viewState    │                          │
│                       │ Manager      │                          │
│                       └──────┬───────┘                          │
│                              │                                   │
│              ┌───────────────┴───────────────┐                  │
│              ▼                               ▼                  │
│     ┌──────────────┐               ┌──────────────┐            │
│     │ Normal View  │               │ Model        │            │
│     │              │               │ Selector     │            │
│     │ - InputBox   │               │              │            │
│     │ - Viewport   │               │ - List       │            │
│     └──────────────┘               │ - Navigation │            │
│                                     └──────────────┘            │
│                                           │                      │
│                                           ▼                      │
│                                    ┌──────────────┐             │
│                                    │ Config      │             │
│                                    │ Manager     │             │
│                                    └──────────────┘             │
└─────────────────────────────────────────────────────────────────┘
         │                                          │
         │ newPrompt() with Meta                   │
         ▼                                          ▼
┌─────────────────┐                    ┌─────────────────┐
│ Agent (ACP)     │                    │ Config File     │
│                 │                    │ ~/.nfa/nfa.json │
│ - Initialize()  │                    │                 │
│ - Prompt()      │                    │ - SaveConfig()  │
└─────────────────┘                    └─────────────────┘
```

### 状态机

```
┌─────────────────────────────────────────────────────────────┐
│                        ChatUI States                         │
└─────────────────────────────────────────────────────────────┘

viewState = viewStateInput
    │
    │ 用户输入 /model[:target]
    │ 按回车
    ▼
viewState = viewStateModelSelect
    │
    │ 用户操作:
    │  - 上下键: 更新选择光标
    │  - ESC: 返回 viewStateInput
    │  - 回车: 应用模型, 保存配置, 返回 viewStateInput
    ▼
viewState = viewStateInput
```

### 数据流

```
启动流程:
1. root.go: 加载 ~/.nfa/nfa.json → cfg
2. root.go: 将 cfgPath 放入 context
3. ChatUI.Run(): 从 context 获取 cfgPath 和 cfg
4. ChatUI.initAgent(): conn.Initialize()
5. Agent.Initialize(): 返回 Meta:
   - availableModels: []string
   - modelDescriptions: map[string]string
6. ChatUI: 存储 availableModels, modelDescriptions
7. ChatUI: selectedModels = cfg.DefaultModels

模型切换流程:
1. 用户输入 /model
2. ChatUI.Update(): 解析命令
3. enterModelSelectMode(ModelTypeMain):
   - 创建 ModelSelector(available, descriptions, selected)
   - viewState = viewStateModelSelect
4. ModelSelector.View(): 渲染菜单
5. 用户按回车
6. ChatUI.Update(): 调用 applyModelAndSave()
7. applyModelAndSave():
   - selectedModels.Main = modelID
   - cfg.DefaultModels = selectedModels
   - SaveConfig(cfgPath, cfg)
8. viewState = viewStateInput
9. 显示成功消息

对话流程:
1. 用户输入问题, 按回车
2. newPrompt():
   - 构造 PromptRequest{
       Meta: {
         "modelName": selectedModels.Main,
         "fastModel": selectedModels.Fast,
         "visionModel": selectedModels.Vision,
       }
     }
3. conn.Prompt(req)
4. Agent.Prompt():
   - 从 Meta 读取 modelName, fastModel, visionModel
   - 覆盖 defaultModels
   - 使用指定模型调用 Genkit
```

## Risks / Trade-offs

### Risk 1: 配置文件保存失败导致数据丢失
**风险**: 磁盘满、权限不足或进程崩溃导致配置无法保存
**缓解**:
- 保存失败时显示明确的错误提示
- 不阻塞 UI 继续运行（用户可以在当前会话继续使用新模型）
- 考虑未来添加"保存到临时文件，下次启动时恢复"的机制

### Risk 2: 并发修改配置文件
**风险**: 多个 NFA 实例同时运行，可能互相覆盖配置
**缓解**:
- 文件锁在 Go 标准库中实现复杂
- 当前不处理（单用户使用场景，同时运行多个实例不常见）
- 未来可以考虑使用文件锁或原子替换

### Risk 3: 模型描述信息与配置文件不同步
**风险**: 用户在配置文件中添加了新的模型描述，但 Agent 仍返回空描述
**缓解**:
- Agent 端在注册模型时从 `ModelConfig.Description` 读取
- 配置文件修改后重启应用即可生效
- 不在运行时热加载配置文件（避免复杂度）

### Trade-off 1: 选择器样式简洁 vs 信息丰富
**决策**: 优先简洁，无边框，描述限制 80 字符
**权衡**: 用户可能看不到完整的模型描述
**理由**: 终端界面宽度有限，简洁样式更符合 TUI 美学

### Trade-off 2: 不验证模型是否可用
**决策**: 直接保存用户选择的模型 ID，不验证是否在 `availableModels` 中
**权衡**: 用户可能配置一个不存在的模型，导致下次对话失败
**理由**:
- 简化实现
- 支持高级用户预配置未来会添加的模型
- Agent 会在模型不存在时返回清晰的错误信息

## Migration Plan

**部署步骤**:
1. 更新 ModelConfig 结构体，添加 `Description` 字段（向后兼容，omitempty）
2. Agent 端返回 `modelDescriptions` Meta
3. UI 端实现 ModelSelector 组件和状态管理
4. 更新配置文件保存逻辑
5. 测试多种场景：交互选择、直接命令、配置文件格式

**回滚策略**:
- 所有更改都是新增功能，不修改现有行为
- 如果出现严重问题，可以回滚到之前的版本
- 配置文件格式保持不变，旧版本可以正常读取

**数据迁移**:
- 无需迁移现有配置文件
- `Description` 字段为可选，旧配置文件不影响使用

## Open Questions

### Q1: 是否需要在配置文件中验证模型名称格式？
**当前设计**: 不验证，直接保存用户输入
**考虑**: 可能导致用户配置 `invalid-format` 然后在对话时才发现错误
**决策**: 保持不验证，在 Agent 返回模型错误时提示用户

### Q2: 模型描述的来源是什么？
**当前设计**: 从 `ModelConfig.Description` 字段读取
**考虑**: 如果用户没有配置描述，是否需要生成默认描述（如从模型名推断）？
**决策**: 暂不生成默认描述，保持为空

### Q3: 是否需要支持搜索/过滤模型列表？
**当前设计**: 不支持，显示所有可用模型
**考虑**: 如果模型列表很长（>20个），用户滚动查找可能不便
**决策**: 暂不实现，观察实际使用情况后再决定

### Q4: 是否需要显示模型的额外信息（价格、上下文长度等）？
**当前设计**: 仅显示描述文本
**考虑**: `ModelConfig` 已有 `Cost`, `ContextWindow` 等字段
**决策**: 暂不显示，未来可以通过 "详情" 视图展示