## ADDED Requirements

### Requirement: Agent 消息 Markdown 渲染

系统 SHALL 使用 glamour 库渲染 Agent 输出的 Markdown 内容，包括标题、列表、代码块、加粗、斜体等格式。

#### Scenario: 渲染标题

- **WHEN** Agent 输出包含 `## 标题` 的 Markdown 内容
- **THEN** 系统将其渲染为格式化的标题样式

#### Scenario: 渲染代码块

- **WHEN** Agent 输出包含 fenced code block 的 Markdown 内容
- **THEN** 系统将其渲染为带语法高亮的代码块

#### Scenario: 渲染失败回退

- **WHEN** glamour 渲染失败
- **THEN** 系统输出原始 Markdown 文本

### Requirement: 终端主题适配

系统 SHALL 自动检测终端深/浅色主题，并使用对应的 glamour 样式渲染 Markdown。

#### Scenario: 深色终端

- **WHEN** 终端为深色主题
- **THEN** 系统使用 dark 样式渲染 Markdown

#### Scenario: 浅色终端

- **WHEN** 终端为浅色主题
- **THEN** 系统使用 light 样式渲染 Markdown

### Requirement: 流式输出实时渲染

系统 SHALL 在流式输出时实时渲染累积的 Markdown 内容。

#### Scenario: 实时渲染不完整 Markdown

- **WHEN** Agent 流式输出 `这是一个 **重`（不完整）
- **THEN** 系统渲染当前累积内容（可能显示原始星号）

#### Scenario: 实时渲染完整 Markdown

- **WHEN** Agent 流式输出 `这是一个 **重要** 的消息`（完整）
- **THEN** 系统渲染格式化的 Markdown（加粗效果）