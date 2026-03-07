## 1. 文件结构准备

- [x] 1.1 创建 `pkg/tools/fs/` 目录
- [x] 1.2 创建 `pkg/tools/fs/read.go` 文件

## 2. 核心 Read 工具实现

- [x] 2.1 定义常量（工具名、最大读取大小 1MB）
- [x] 2.2 定义 `ReadInput` 结构体（Path、Offset、Limit 字段）
- [x] 2.3 定义 `ReadOutput` 结构体（Content、Size、BytesRead、Truncated 字段）
- [x] 2.4 定义 `FileReader` 结构体
- [x] 2.5 实现 `NewFileReader()` 构造函数
- [x] 2.6 实现 `DefineReadTool()` 方法，使用 `genkit.DefineTool()` 注册工具
- [x] 2.7 实现输入验证逻辑（path 不为空、offset >= 0、limit <= 1MB）
- [x] 2.8 实现文件打开和错误处理（`os.Open()`）
- [x] 2.9 实现文件信息获取（`file.Stat()`）
- [x] 2.10 实现 offset 定位（`file.Seek()`）
- [x] 2.11 实现限制读取逻辑（`io.LimitReader()`，limit 为 0 时使用默认值 1MB）
- [x] 2.12 实现截断检测（尝试多读 1 字节确认）
- [x] 2.13 实现工具描述文本（中英文）

## 3. 工具注册

- [x] 3.1 在 `pkg/agents/genkit.go` 中导入 `github.com/yhlooo/nfa/pkg/tools/fs`
- [x] 3.2 在 `InitGenkit()` 方法中创建 `FileReader` 实例
- [x] 3.3 在 `InitGenkit()` 方法中将 Read 工具添加到 `availableTools` 数组
- [x] 3.4 确保工具在日志中正确输出

## 4. 代码质量检查

- [x] 4.1 运行 `go fmt ./...` 格式化代码
- [x] 4.2 运行 `go vet ./...` 检查代码问题
- [x] 4.3 运行 `go test ./...` 确保现有测试通过
- [x] 4.4 手动测试 Read 工具功能（可选）

## 5. 验证收尾

- [x] 5.1 确认所有文件已正确创建和修改
- [x] 5.2 确认代码符合 Go 语言规范
- [x] 5.3 确认工具可在 Agent 中正常调用
