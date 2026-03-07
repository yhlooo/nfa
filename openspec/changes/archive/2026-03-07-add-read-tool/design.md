## Context

NFA Agent 使用 Firebase Genkit 框架进行 AI 编排，现有工具主要包括 WebSearch、WebBrowse、AlphaVantage 和 Skill。这些工具都通过网络或特定服务获取数据，但缺少直接访问本地文件系统的能力。

当前工具架构：
- 工具定义在 `pkg/tools/` 下的各个子目录中
- 通过 `genkit.DefineTool()` 注册工具
- 工具函数签名：`func(ctx *ai.ToolContext, input Input) (Output, error)`
- 在 `pkg/agents/genkit.go` 的 `InitGenkit()` 方法中注册所有工具

## Goals / Non-Goals

**Goals:**
- 提供本地文件读取能力，支持任意路径的文件访问
- 支持按字节偏移读取，便于处理大文件
- 内置 1MB 大小限制，防止内存问题
- 遵循现有工具架构模式，保持代码一致性

**Non-Goals:**
- 暂不实施路径安全限制（允许读取任意路径）
- 不支持特殊文件格式转换（如 PDF、图片），保持简单
- 不实现文件写入、列表等其他文件系统操作

## Decisions

### 1. 包结构：`pkg/tools/fs/`

选择创建 `fs` 包而不是直接放在 `tools` 根目录下：
- **理由**：`fs` 为未来可能的其他文件系统操作（Write、List、Delete 等）预留空间
- **权衡**：多一层目录，但组织更清晰

### 2. 使用 `io.LimitReader` 实现大小限制

选择 `io.LimitReader` 而非手动切片：
- **理由**：标准库实现，简洁可靠，自动处理读取循环
- **权衡**：需要在读取后额外检测是否真正到达文件末尾

### 3. 截断检测策略

通过尝试多读 1 字节来确认是否截断：
- **理由**：`io.LimitReader` 在达到限制后返回 EOF，无法区分"刚好读完"和"被截断"
- **权衡**：增加一次系统调用，但结果准确

### 4. 错误处理

使用 Go 标准的 error 返回值，不在结构体中包含错误字段：
- **理由**：符合 Go 语言惯例，与现有工具一致
- **权衡**：无

### 5. limit 参数语义

limit 为 0 时表示使用默认值（1MB），最大值不超过 1MB：
- **理由**：提供合理默认值，同时允许灵活配置
- **权衡**：需要额外验证逻辑

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 读取大文件导致内存压力 | 1MB 硬限制，使用 `io.LimitReader` |
| 路径遍历攻击 | 暂不限制（用户明确要求），后续可添加白名单 |
| 二进制文件内容不可读 | 返回原始字节，由 Agent 自行处理 |
| offset 超出文件大小 | 使用 `Seek()` 处理，返回 EOF |
| 并发读取同一文件 | `os.Open()` 每次创建新的文件句柄，安全 |

## Migration Plan

**部署步骤：**
1. 创建 `pkg/tools/fs/read.go` 文件
2. 实现 `FileReader` 结构体和 `DefineReadTool()` 方法
3. 在 `pkg/agents/genkit.go` 中导入 `fs` 包并注册工具
4. 运行 `go fmt ./...`、`go vet ./...`、`go test ./...` 验证

**回滚策略：**
- 移除 `pkg/tools/fs/` 目录
- 从 `pkg/agents/genkit.go` 中移除相关代码

## Open Questions

无。需求清晰，可直接实现。
