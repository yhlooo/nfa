# Design: Add Non-Interactive Mode

## Context

nfa 当前采用交互式终端模式，架构如下：

```
nfa (cobra) → NFAAgent ←→ ChatUI (Bubble Tea TUI)
                ↑________↑
                  ACP Protocol
```

- **NFAAgent**: 实现业务逻辑，通过 ACP 协议与客户端通信
- **ChatUI**: 实现交互式 TUI，使用 Bubble Tea 框架，处理用户输入并显示响应
- **ACP Protocol**: Agent Communication Protocol，用于 Agent 和 Client 之间通信

当前流程：
1. 用户启动 `nfa` 进入交互式终端
2. ChatUI 创建 session，等待用户输入
3. 用户输入问题 → ChatUI 通过 ACP 发送到 NFAAgent
4. NFAAgent 流式返回响应 → ChatUI 显示
5. 循环等待下一个问题

约束：
- 复用现有 ACP 通信机制
- 复用现有 ChatUI 的渲染逻辑（包括思考过程、工具调用等）
- 不影响现有的交互式模式

## Goals / Non-Goals

**Goals:**
1. 支持通过位置参数传递初始问题
2. 支持通过 `-p` 标志在回答后自动退出
3. 三种使用模式：
   - `nfa` - 交互式模式（无变化）
   - `nfa '问题'` - 交互式模式 + 自动发送初始问题
   - `nfa '问题' -p` - 非交互式单次问答模式
4. 非交互式模式下的输出与交互式模式一致（包含思考过程、工具调用等）

**Non-Goals:**
1. 不修改 ACP 协议
2. 不修改 NFAAgent 核心逻辑
3. 不创建新的 UI 组件

## Decisions

### Decision 1: 复用 ChatUI 而非创建新的客户端

**Rationale:**
- ChatUI 已经实现了完整的消息渲染逻辑（思考过程、工具调用、错误处理等）
- 非交互式模式的输出要求与交互式模式一致
- 避免代码重复和维护成本

**Implementation:**
- 在 ChatUI 中添加 `InitialPrompt` 和 `AutoExitAfterResponse` 选项
- 在 `Update` 处理 `PromptResponse` 时，如果 `AutoExitAfterResponse` 为 true，返回 `tea.Quit`

**Alternatives Considered:**
- 创建新的 SimpleClient 直接调用 Agent：需要重新实现响应收集逻辑，代码重复

### Decision 2: 初始问题在 Init 期间发送

**Rationale:**
- 保持消息流的一致性，初始问题通过相同的 `newPrompt` 命令发送
- 在 Init 的 `tea.Sequence` 中加入发送初始问题的命令，确保在 Session 创建之后发送

**Implementation:**
```go
func (ui *ChatUI) Init() tea.Cmd {
    cmds := []tea.Cmd{
        ui.newSession(),
        ui.printHello(),
        textarea.Blink,
    }
    if ui.initialPrompt != "" {
        cmds = append(cmds, func() tea.Msg {
            // 使用相同的逻辑发送 PromptRequest
            return acp.PromptRequest{...}
        })
    }
    return tea.Sequence(cmds...)
}
```

**Alternatives Considered:**
- 在 Run 方法中直接发送：绕过了 Tea 的事件循环，可能导致状态不一致

### Decision 3: 自动退出在 PromptResponse 处理时触发

**Rationale:**
- PromptResponse 标志着一轮对话的完成
- 此时所有流式输出已经发送到 UI
- 退出时机准确，不会提前或延迟

**Implementation:**
```go
case acp.PromptResponse:
    ui.modelUsage = agents.GetMetaCurrentModelUsageValue(typedMsg.Meta)
    cmds = append(cmds, ui.vp.Flush())
    if ui.autoExitAfterResponse {
        cmds = append(cmds, tea.Quit)
    }
```

**Alternatives Considered:**
- 在 SessionUpdate 的最后一次更新时退出：难以判断何时是最后一次更新

### Decision 4: -p 标志命名为 --print

**Rationale:**
- `-p` 简短易记
- `--print` 自解释，清晰表达功能意图

## Risks / Trade-offs

### Risk 1: 自动退出时流式输出可能不完整

**Mitigation:** 在 `PromptResponse` 时触发退出，此时所有流式输出已经处理完毕。`vp.Flush()` 确保所有缓存的消息都已输出。

### Risk 2: 初始问题发送时机可能出错

**Mitigation:** 使用 `tea.Sequence` 确保在 `newSession` 完成后再发送初始问题。

### Risk 3: 非交互式模式下错误处理不一致

**Mitigation:** 错误处理逻辑保持不变，只是退出时机不同。错误会正常显示到 UI，然后触发退出。

### Risk 4: -p 模式下欢迎信息的显示

**Trade-off:** 保留欢迎信息可以让用户知道正在使用的 model，但会增加输出。保持显示与交互式模式一致。

## Migration Plan

1. 修改 `pkg/commands/root.go`：添加 `-p` flag，修改 `RunE` 逻辑
2. 修改 `pkg/ui/chat/ui.go`：添加 `InitialPrompt` 和 `AutoExitAfterResponse` 选项
3. 修改 `pkg/ui/chat/ui.go`：修改 `Init` 和 `Update` 方法
4. 测试三种模式的行为
5. 更新文档（可选）

无特殊部署步骤，无数据库迁移，无外部依赖变更。

**Rollback Strategy:** 删除 `-p` flag 和相关代码，恢复原有行为。

## Open Questions

无。
