## Why

NFA Agent 目前缺少读取本地文件的能力。在金融分析场景中，用户经常需要让 Agent 分析本地的数据文件（如 CSV、JSON、配置文件等），但现有工具（WebFetch）只支持网络资源，无法直接访问本地文件系统。

## What Changes

- 新增 `Read` 工具，允许 Agent 读取本地文件系统中的任意文件
- 工具支持按字节偏移读取（`offset` + `limit`），便于处理大文件
- 内置 1MB 大小限制，超出部分自动截断，防止内存问题
- 返回文件内容、大小、实际读取字节数和截断标志
- 读取错误通过标准 error 返回值处理

## Capabilities

### New Capabilities
- `fs-read`: 本地文件系统读取能力，支持读取任意路径的文件，提供偏移和大小限制功能

### Modified Capabilities
无

## Impact

**新增文件：**
- `pkg/tools/fs/read.go` - Read 工具实现

**修改文件：**
- `pkg/agents/genkit.go` - 注册 Read 工具到 Genkit

**依赖：**
- 无新增外部依赖，使用 Go 标准库（`os`、`io`）

**系统影响：**
- Agent 获得访问本地文件的能力，可用于读取用户数据、配置文件等
- 暂不实施安全限制（允许读取任意路径），后续可根据需要添加
