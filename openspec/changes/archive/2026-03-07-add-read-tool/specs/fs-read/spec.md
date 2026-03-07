## ADDED Requirements

### Requirement: 工具可读取本地文件

系统 SHALL 提供 `Read` 工具，允许 Agent 读取本地文件系统中的任意文件。工具 SHALL 接受文件路径作为输入，并返回文件内容。

#### Scenario: 成功读取文本文件
- **WHEN** Agent 调用 Read 工具，提供有效的文本文件路径
- **THEN** 系统返回文件内容、文件大小和实际读取的字节数

#### Scenario: 读取不存在的文件
- **WHEN** Agent 调用 Read 工具，提供不存在的文件路径
- **THEN** 系统返回错误，说明文件不存在

#### Scenario: 读取二进制文件
- **WHEN** Agent 调用 Read 工具，读取二进制文件
- **THEN** 系统返回原始字节内容（可能不可读）

### Requirement: 支持字节偏移读取

系统 SHALL 支持 `offset` 参数，允许从指定字节位置开始读取文件。

#### Scenario: 从文件开头读取
- **WHEN** Agent 调用 Read 工具，offset 为 0 或未指定
- **THEN** 系统从文件开头开始读取

#### Scenario: 从文件中间读取
- **WHEN** Agent 调用 Read 工具，offset 为有效正数
- **THEN** 系统从指定字节位置开始读取

#### Scenario: offset 超出文件大小
- **WHEN** Agent 调用 Read 工具，offset 大于文件大小
- **THEN** 系统返回空内容和文件大小信息

### Requirement: 支持读取大小限制

系统 SHALL 支持 `limit` 参数，限制单次读取的最大字节数。系统 SHALL 使用 `io.LimitReader` 实现大小限制。

#### Scenario: 使用默认限制
- **WHEN** Agent 调用 Read 工具，limit 为 0 或未指定
- **THEN** 系统使用默认 1MB 限制

#### Scenario: 使用自定义限制
- **WHEN** Agent 调用 Read 工具，limit 为有效正数且不超过 1MB
- **THEN** 系统使用指定的限制值

#### Scenario: limit 超过最大值
- **WHEN** Agent 调用 Read 工具，limit 大于 1MB
- **THEN** 系统返回验证错误

### Requirement: 检测文件截断

系统 SHALL 检测文件是否被截断（即文件还有更多内容未被读取），并在返回结果中设置 `Truncated` 标志。

#### Scenario: 文件未截断
- **WHEN** 文件大小小于或等于读取限制
- **THEN** `Truncated` 为 false

#### Scenario: 文件被截断
- **WHEN** 文件大小大于读取限制
- **THEN** `Truncated` 为 true，且 `BytesRead` 等于限制值

### Requirement: 错误处理

系统 SHALL 通过标准 error 返回值返回读取错误，不在输出结构体中包含错误字段。

#### Scenario: 文件权限不足
- **WHEN** Agent 尝试读取无权限访问的文件
- **THEN** 系统返回权限错误

#### Scenario: 路径为空
- **WHEN** Agent 调用 Read 工具，path 为空字符串
- **THEN** 系统返回验证错误

### Requirement: 工具注册

系统 SHALL 将 Read 工具注册到 Genkit，使其可被 Agent 调用。工具名称 SHALL 为 `Read`。

#### Scenario: 工具注册
- **WHEN** 系统初始化 Genkit
- **THEN** Read 工具被注册，可通过 `ai.WithTools()` 使用
