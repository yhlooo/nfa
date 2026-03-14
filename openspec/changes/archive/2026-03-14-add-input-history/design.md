## Context

**当前状态**
NFA 系统使用 Bubble Tea 框架实现交互式终端 UI：
1. `InputBox` 组件使用 `textarea.Model` 处理用户输入
2. 用户按 Enter 提交输入，内容被发送给 Agent
3. 没有历史记录功能

**架构约束**
- 使用 Bubble Tea 的 `tea.Msg` 模式处理事件
- 配置和数据存储在 `~/.nfa/` 目录
- 现有模式：配置使用 JSON 文件存储（参考 `pkg/configs/`）

## Goals / Non-Goals

**Goals:**
- 记录用户历史输入到持久化文件
- 支持上下键浏览历史（单行模式）
- 限制历史记录数量，避免文件过大

**Non-Goals:**
- 不支持历史搜索功能
- 不支持历史编辑功能
- 不显示历史索引提示
- 不新增外部依赖

## Decisions

### 1. 历史存储格式：JSON 数组

**决策**: 使用 JSON 数组存储历史记录，每条记录包含 `ts` 和 `content`。

```json
[
  {"ts": 1709500000000000000, "content": "message1"},
  {"ts": 1709500000000000001, "content": "message2"}
]
```

**理由**:
- **简单**: JSON 格式易于读写和调试
- **有序**: 数组天然保持时间顺序
- **兼容**: 与现有配置文件格式一致

### 2. 上下键行为：仅单行模式启用

**决策**: 只有在单行输入模式下，↑/↓ 键才用于浏览历史。多行模式下保留原有光标移动行为。

**理由**:
- **不破坏现有功能**: 多行模式下用户需要上下移动光标
- **简单明确**: 通过 `InputBox.multiLine` 字段判断模式
- **符合预期**: 类似 readline 的行为

### 3. 历史导航状态管理

**决策**: 在 `InputBox` 中维护 `historyIndex` 和 `historyTempValue` 状态。

```
historyIndex:
  - -1: 未在浏览历史（正在输入新内容）
  - 0: 最新的历史记录
  - n-1: 最旧的历史记录

historyTempValue:
  - 用户开始浏览历史时保存当前输入
  - 按 ↓ 到底部时恢复
```

**理由**:
- **直觉体验**: 与 shell 历史行为一致
- **不丢失输入**: 用户输入一半时浏览历史，回来能恢复

### 4. 保存时机：每次提交后立即保存

**决策**: 用户每次提交非空输入后，立即追加到历史并保存文件。

**理由**:
- **不丢失数据**: 即使程序异常退出，历史也已保存
- **简单实现**: 无需复杂的退出处理逻辑

### 5. 文件路径：~/.nfa/history.json

**决策**: 历史文件存储在 `~/.nfa/history.json`，与配置文件同目录。

**理由**:
- **一致性**: 与现有数据文件位置一致
- **易于管理**: 用户可以轻松找到和备份

## Architecture

### 组件结构

```
pkg/history/
├── history.go      # History 结构体和方法
└── load.go         # Load/Save 函数

pkg/ui/chat/input.go
├── InputBox 新增字段:
│   ├── history           *history.History
│   ├── historyPath       string
│   ├── historyIndex      int
│   └── historyTempValue  string
└── Update() 新增逻辑:
    ├── 处理 ↑ 键: history.Up()
    └── 处理 ↓ 键: history.Down()
```

### 数据流

```
ChatUI.Run()
    │
    ├── LoadHistory(path)  // 初始化时加载
    │
    └── 创建 InputBox{history, historyPath}

用户按 Enter 提交
    │
    ├── content = input.Value()
    │
    ├── if content != "":
    │       history.Add(content)
    │       SaveHistory(path, history)
    │
    └── 发送给 Agent

用户按 ↑ (单行模式)
    │
    ├── if historyIndex == -1:
    │       historyTempValue = input.Value()  // 保存当前输入
    │
    └── value := history.Up()
            input.SetValue(value)

用户按 ↓ (单行模式)
    │
    └── value := history.Down()
            if value == "":
                input.SetValue(historyTempValue)  // 恢复
            else:
                input.SetValue(value)
```

### History 结构体设计

```go
type Entry struct {
    TS      int64  `json:"ts"`      // Unix timestamp (nano)
    Content string `json:"content"`
}

type History struct {
    entries   []Entry
    maxLen    int      // 默认 100
    // 导航状态（不持久化）
    navIndex  int      // 当前浏览位置
}
```

## Risks / Trade-offs

### Risk 1: 历史文件损坏

**风险**: 如果程序在写入过程中崩溃，可能导致文件损坏。

**缓解**:
- 文件较小（100 条记录），写入速度快
- 损坏时可删除文件，程序会自动重新创建

### Risk 2: 并发写入

**风险**: 多个 NFA 实例同时运行可能导致历史冲突。

**缓解**:
- 当前场景预期单实例运行
- 文件写入是原子操作（小文件）
- 未来可考虑文件锁

### Trade-off: 不去重

**决策**: 不对重复输入去重。

**理由**:
- 保持实现简单
- 重复输入本身反映了使用模式
- 100 条限制已足够控制大小

## Open Questions

### Q1: 是否需要限制单条记录的长度？

**当前设计**: 不限制

**决策**: 暂不限制，观察实际使用情况。如果出现异常长的输入，可后续添加截断逻辑。

### Q2: 历史记录是否包含斜杠命令？

**当前设计**: 包含所有非空输入

**决策**: 记录所有输入，包括 `/model`、`/clear` 等命令。这与 shell 行为一致。
