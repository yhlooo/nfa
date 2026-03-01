# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

NFA (Not Financial Advice，非财务建议) 是一个基于 Go 语言构建的金融交易 LLM AI Agent。它使用 Firebase Genkit 进行 AI 编排，支持多种模型提供商（Ollama、DeepSeek、OpenAI 兼容等）。Agent 通过基于主题的路由处理用户查询，并执行存储在 Markdown 文件中的自定义技能。

## 构建和运行

```bash
# 构建并运行
go run ./cmd/nfa

# 使用单行提示词运行
go run ./cmd/nfa "分析一下特斯拉"

# 使用详细日志运行
go run ./cmd/nfa -v "问题"

# 以打印后退出模式运行（非交互式）
go run ./cmd/nfa -p "单次问题"

# 列出可用模型
go run ./cmd/nfa models list

# 列出内部工具
go run ./cmd/nfa tools
```

## 配置文件

配置存储在 `~/.nfa/nfa.json`。核心结构：

- `modelProviders` - 模型提供商配置数组（Ollama、DeepSeek、OpenAI 兼容、阿里云、智谱）
- `defaultModels` - 不同任务的默认模型（primary 主模型、light 轻量模型、vision 视觉模型）
- `dataProviders` - 数据提供商配置（AlphaVantage、腾讯云）

详细配置选项参见 `docs/reference/config.md`。

## 代码质量检查

编辑代码后，必须按以下顺序执行检查：

```bash
# 1. 代码格式化（必须在编辑后立即执行）
go fmt ./...

# 2. 静态分析检查语法问题（使用 go vet 而非 go build）
go vet ./...

# 3. 运行单元测试确认功能正常
go test ./...
```

**重要说明**：
- `go fmt ./...` - 自动格式化所有 Go 代码，确保代码风格一致
- `go vet ./...` - 检查代码中的常见错误，而非 `go build`（vet 能发现 build 无法检测的问题）
- `go test ./...` - 运行所有单元测试，确保修改没有破坏现有功能

## 架构设计

### 核心组件

**入口点** (`cmd/nfa/main.go`):
- 设置 SIGINT/SIGTERM 信号处理
- 创建并执行根 cobra 命令

**命令层** (`pkg/commands/`):
- `root.go` - 根命令，包含全局选项（日志级别、数据根目录）和运行时选项（模型选择）
- `models.go` - 模型命令，列出可用模型
- 处理从 `~/.nfa/nfa.json` 加载配置

**Agent 系统** (`pkg/agents/`):
- `agent.go` - 核心 NFA Agent，包含会话管理、模型路由和 ACP（Agent Communication Protocol）集成
- `genkit.go` - Genkit AI 框架初始化
- `prompts.go` - Agent 的系统提示词
- `flows/` - AI 流程实现：
  - `topic_routing.go` - 将用户查询路由到适当的主题（Query 查询、StockAnalysis 个股分析、PortfolioAnalysis 组合分析等）
  - `chats.go` - 多 Agent 聊天编排
  - `summarize.go` - 对话摘要生成
  - `chat_simple.go` - 简单单 Agent 聊天流程

**模型管理** (`pkg/models/`):
- `providers.go` - 模型提供商接口和实现
- `model_routing.go` - Models 结构体，用于 primary/light/vision 模型选择
- 各提供商文件：`ollama.go`、`deepseek.go`、`openai_compatible.go`、`aliyun_dashscope.go`、`zhipu_bigmodel.go`

**技能系统** (`pkg/skills/`):
- `skill_loader.go` - 从 `~/.nfa/skills/` 目录加载技能
- `skill_parser.go` - 解析带 YAML frontmatter 的 SKILL.md 文件
- `skill_tool.go` - 将技能作为工具暴露给 Agent
- 技能是存储为 Markdown 文件的用户定义工作流

**工具** (`pkg/tools/`):
- `websearch/` - 网络搜索工具（腾讯云 WSA）
- `webbrowse/` - 通过 ChromeDP 进行网页浏览
- `alphavantage/` - Alpha Vantage 金融数据集成
- `generic.go` - 通用工具定义

**用户界面** (`pkg/ui/chat/`):
- 基于 Bubble Tea 的 TUI 交互式聊天界面
- `model_selector.go` - 交互式模型选择菜单
- `acp.go` - ACP 客户端连接处理

### 主题路由

Agent 使用主题路由来分类用户查询：
- **Query** - 简单信息查询
- **StockAnalysis** - 个股分析
- **PortfolioAnalysis** - 投资组合分析
- **ShortTermTrendForecast** - 短期趋势预测
- **Basic** - 基础金融知识
- **Comprehensive** - 复杂的多方面问题
- **Others** - 其他主题

### 技能系统

技能存储在 `~/.nfa/skills/<技能名>/SKILL.md`，包含 YAML frontmatter：

```markdown
---
name: skill-name
description: 技能描述
---

Agent 需要遵循的分步指令。
```

当 Agent 需要某个技能时，它会调用 `Skill` 工具并传入技能名称，获取指令后按照指令执行。

### 交互式命令

- `/model` - 切换模型（交互式或直接指定：`/model ollama/llama3.2`）
- `/model :light` - 切换轻量模型
- `/model :vision` - 切换视觉模型
- `/clear` - 清空对话上下文
- `/summarize` - 生成对话摘要
- `/exit` - 退出程序

完整命令参考见 `docs/guides/command-line.md`。

## 核心模式

1. **模型路由** - 系统根据任务复杂度将请求路由到不同的模型（primary 主模型 vs light 轻量模型 vs vision 视觉模型）

2. **基于主题的流程** - 用户查询首先经过主题分类，然后选择适当的流程

3. **技能即工具** - 自定义技能通过 `Skill` 工具暴露给 Agent，实现动态行为

4. **会话管理** - 每个用户会话在 `NFAAgent.sessions` 映射中维护对话历史

5. **ACP 协议** - 通过 `acp-go-sdk` 进行 Agent 通信，实现结构化的客户端-服务器交互

6. **配置层级** - 命令行参数 > 配置文件 > 默认值

## 文件位置

- 配置文件：`~/.nfa/nfa.json`
- 日志文件：`~/.nfa/nfa.log`
- 技能目录：`~/.nfa/skills/`
- 数据目录：`~/.nfa/`

## 开发注意事项

- Go 版本：1.24.7
- 使用 Cobra 构建 CLI，Bubble Tea 构建 TUI，Firebase Genkit 进行 AI 编排
- 通过 logr/logrus 记录日志，支持日志轮转（最大 500MB，保留 3 个备份，28 天）
- 通过 `SSLKEYLOGFILE` 环境变量支持 TLS key logging
- ACP 连接使用 io.Pipe 进行双向通信
