# Input History

## Overview

用户历史输入记录功能，支持持久化存储和上下键浏览。

## Requirements

### REQ-001: 历史记录存储
系统应将用户输入存储到 `~/.nfa/history.json` 文件中。

### REQ-002: 存储格式
历史记录应以 JSON 数组格式存储，每条记录包含：
- `ts`: Unix 时间戳（纳秒精度）
- `content`: 输入内容

### REQ-003: 记录数量限制
系统应保留最近 100 条历史记录。

### REQ-004: 空输入过滤
系统不应记录空输入（仅包含空白字符的输入）。

### REQ-005: 上下键浏览
在单行输入模式下：
- ↑ 键应显示更早的历史记录
- ↓ 键应显示更新的历史记录
- 到达最新记录后继续按 ↓ 应恢复用户当前输入

### REQ-006: 多行模式行为
在多行输入模式下，↑/↓ 键应保留原有光标移动行为，不触发历史浏览。

### REQ-007: 实时保存
用户提交输入后应立即保存历史记录到文件。

## Data Model

```json
[
  {"ts": 1709500000000000000, "content": "分析一下特斯拉"},
  {"ts": 1709500000000000001, "content": "/model ollama/llama3.2"},
  {"ts": 1709500000000000002, "content": "什么是市盈率"}
]
```

## API

### LoadHistory(path string) (*History, error)
从指定路径加载历史记录。如果文件不存在，返回空的 History。

### SaveHistory(path string, h *History) error
将历史记录保存到指定路径。

### (h *History) Add(content string)
添加一条新记录。如果超过最大数量，移除最旧的记录。

### (h *History) Up() string
向上浏览（获取更早的记录）。返回空字符串表示已到达最旧。

### (h *History) Down() string
向下浏览（获取更新的记录）。返回空字符串表示已到达最新。
