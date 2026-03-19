# 任务清单

## 实现任务

### 任务 1: 创建 MiniMax 提供商实现文件 ✓

**文件**: `pkg/models/minimax.go`

**内容**:
- 定义 `MinimaxProviderName = "minimax"` 常量
- 定义 `MinimaxBaseURL = "https://api.minimax.chat/v1"` 常量
- 实现 `MinimaxModels()` 函数返回预设模型列表
- 实现 `MinimaxOptions` 结构体
- 实现 `Complete()`, `Plugin()`, `RegisterModels()` 方法

---

### 任务 2: 更新 ModelProvider 结构体 ✓

**文件**: `pkg/models/providers.go`

**变更**: 在 `ModelProvider` 结构体中添加字段：
```go
Minimax *MinimaxOptions `json:"minimax,omitempty"`
```

---

### 任务 3: 添加模型描述国际化消息 ✓

**文件**: `pkg/models/i18n.go`

**变更**: 添加消息定义：
- `MsgModelDescMinimaxM25`
- `MsgModelDescMinimaxM27`

---

### 任务 4: 更新 Genkit 初始化逻辑 ✓

**文件**: `pkg/agents/genkit.go`

**变更**:
1. 在确定插件的 switch 中添加 `case p.Minimax != nil`
2. 在注册模型的 switch 中添加 `case p.Minimax != nil`

---

### 任务 5: 更新配置文档 ✓

**文件**: `docs/reference/config.md`

**变更**: 添加 MiniMax 配置说明章节

---

## 验证任务

### 任务 6: 代码质量检查

运行：
```bash
go fmt ./...
go vet ./...
go test ./...
```

---

### 任务 7: 手动测试

1. 配置 MiniMax API Key
2. 运行 `go run ./cmd/nfa` 验证模型注册
3. 测试模型调用
