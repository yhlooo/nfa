# Memory Storage

## Purpose

长期记忆的持久化存储和读取，包括 MEMORY.md 文件的格式定义、加载、写入，以及在系统提示词中的注入。

## Requirements

### Requirement: 记忆文件加载

系统 SHALL 在程序启动时从 `~/.nfa/MEMORY.md` 加载记忆内容。如果文件不存在，系统 SHALL 以空记忆继续运行，不报错。

#### Scenario: 记忆文件存在时加载

- **WHEN** `~/.nfa/MEMORY.md` 文件存在且内容可读
- **THEN** 系统读取文件全部内容并将其作为记忆内容缓存到内存中

#### Scenario: 记忆文件不存在时启动

- **WHEN** `~/.nfa/MEMORY.md` 文件不存在
- **THEN** 系统以空字符串作为记忆内容，正常启动，不报错

### Requirement: 记忆内容注入系统提示词

系统 SHALL 在每轮对话中将已加载的记忆内容注入系统提示词。记忆内容 SHALL 作为系统提示词中独立的"## 用户记忆"章节呈现。

#### Scenario: 有记忆时的系统提示词

- **WHEN** 记忆内容非空
- **THEN** 系统提示词末尾包含"## 用户记忆"章节，内容为记忆文件的完整文本

#### Scenario: 无记忆时的系统提示词

- **WHEN** 记忆内容为空
- **THEN** 系统提示词中不包含"## 用户记忆"章节

### Requirement: 记忆文件更新

系统 SHALL 在异步总结完成后将新的记忆内容写入 `~/.nfa/MEMORY.md`。写入操作 SHALL 使用 `os.WriteFile` 配合 `0644` 权限。

#### Scenario: 总结完成后写入文件

- **WHEN** 异步总结成功生成新的记忆内容
- **THEN** 系统将内容写入 `~/.nfa/MEMORY.md`，并更新内存中的缓存

#### Scenario: 目录不存在时自动创建

- **WHEN** `~/.nfa/` 目录不存在
- **THEN** 系统使用 `os.MkdirAll` 创建目录后再写入文件

### Requirement: 记忆内容线程安全

系统 SHALL 保证记忆内容的读写操作线程安全。多个 goroutine 可能同时读取记忆内容（系统提示词构建），同时可能有一个 goroutine 更新记忆（异步总结）。

#### Scenario: 并发读写安全

- **WHEN** 一个 goroutine 正在更新记忆内容，同时另一个 goroutine 正在读取记忆内容构建系统提示词
- **THEN** 读取操作不会 panic 或读到不完整的数据，返回更新前或更新后的完整内容
