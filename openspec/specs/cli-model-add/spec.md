# cli-model-add

## Purpose

为 `nfa models` 命令提供 `add` 子命令，允许用户通过命令行交互式地添加或覆盖模型供应商配置。

## Requirements

### Requirement: `models add` 子命令

系统 SHALL 为 `nfa models` 命令提供 `add` 子命令，语法为 `nfa models add <provider> [flags]`，用于添加或覆盖模型供应商配置。

#### Scenario: 子命令注册

- **WHEN** 用户执行 `nfa models --help`
- **THEN** 系统 SHALL 在帮助信息中显示 `add` 子命令
- **AND** `add` 子命令描述为"Add a model provider configuration"

#### Scenario: 添加 DeepSeek 供应商

- **WHEN** 用户执行 `nfa models add deepseek --apiKey sk-xxxxxxxx`
- **THEN** 系统 SHALL 在配置文件的 `modelProviders` 数组中追加或覆盖一个 `deepseek` 配置项
- **AND** 配置项的 `apiKey` 字段 SHALL 为 `sk-xxxxxxxx`
- **AND** 配置项的 `baseURL` 字段 SHALL 留空（使用默认值）

#### Scenario: 添加 Qwen 供应商并指定 baseURL

- **WHEN** 用户执行 `nfa models add qwen --apiKey sk-xxx --baseURL https://custom.endpoint/v1`
- **THEN** 系统 SHALL 新建 `qwen` 配置项
- **AND** `baseURL` SHALL 为 `https://custom.endpoint/v1`

#### Scenario: 添加 Moonshot 供应商

- **WHEN** 用户执行 `nfa models add moonshot --apiKey sk-xxx`
- **THEN** 系统 SHALL 新建 `moonshotai` 配置项

#### Scenario: 添加 MiniMax 供应商

- **WHEN** 用户执行 `nfa models add minimax --apiKey sk-xxx`
- **THEN** 系统 SHALL 新建 `minimax` 配置项

#### Scenario: 添加 OpenRouter 供应商

- **WHEN** 用户执行 `nfa models add openrouter --apiKey sk-xxx`
- **THEN** 系统 SHALL 新建 `openrouter` 配置项

#### Scenario: 添加智谱 AI 供应商

- **WHEN** 用户执行 `nfa models add z-ai --apiKey xxx`
- **THEN** 系统 SHALL 新建 `z-ai` 配置项

#### Scenario: 添加 OpenAI 兼容供应商

- **WHEN** 用户执行 `nfa models add openai-compatible --name my-provider --apiKey sk-xxx --baseURL https://api.example.com/v1`
- **THEN** 系统 SHALL 新建 `openaiCompatible` 配置项
- **AND** `name` 字段 SHALL 为 `my-provider`

#### Scenario: 添加 Ollama 供应商

- **WHEN** 用户执行 `nfa models add ollama --serverAddress http://192.168.1.100:11434 --timeout 600`
- **THEN** 系统 SHALL 新建 `ollama` 配置项
- **AND** `serverAddress` SHALL 为 `http://192.168.1.100:11434`
- **AND** `timeout` SHALL 为 `600`

#### Scenario: Ollama 使用默认值

- **WHEN** 用户执行 `nfa models add ollama` 且不指定任何 flag
- **THEN** 系统 SHALL 新建 `ollama` 配置项
- **AND** `serverAddress` SHALL 留空（使用默认值）

### Requirement: 供应商名映射

系统 SHALL 将用户输入的供应商名映射到对应的配置 JSON 字段名。

#### Scenario: 已知供应商名映射

- **WHEN** 用户输入 `moonshot`
- **THEN** 系统 SHALL 映射为 JSON 字段 `moonshotai`

#### Scenario: 未知供应商名

- **WHEN** 用户输入一个不在支持的供应商列表中的名称（如 `unknown-provider`）
- **THEN** 系统 SHALL 返回错误，提示"未知的供应商类型"
- **AND** 错误信息中 SHALL 列出所有支持的供应商名

### Requirement: 必填参数校验

系统 SHALL 根据供应商类型校验必填参数。

#### Scenario: DeepSeek 未提供 apiKey

- **WHEN** 用户执行 `nfa models add deepseek` 且不指定 `--apiKey`
- **THEN** 系统 SHALL 返回错误，提示缺少必填参数 `--apiKey`

#### Scenario: OpenAI 兼容供应商未提供 name

- **WHEN** 用户执行 `nfa models add openai-compatible --apiKey sk-xxx` 且不指定 `--name`
- **THEN** 系统 SHALL 返回错误，提示缺少必填参数 `--name`

#### Scenario: Ollama 未提供任何参数

- **WHEN** 用户执行 `nfa models add ollama` 且不指定任何 flag
- **THEN** 系统 SHALL 成功创建配置（Ollama 无必填参数）

### Requirement: 覆盖已有配置

若同类型供应商已存在配置，系统 SHALL 用新配置覆盖而非追加。

#### Scenario: 覆盖已有的 DeepSeek 配置

- **WHEN** 配置文件中已存在 `deepseek` 配置项（`apiKey` 为 `sk-old`）
- **AND** 用户执行 `nfa models add deepseek --apiKey sk-new`
- **THEN** 系统 SHALL 将已有 `deepseek` 配置项的 `apiKey` 更新为 `sk-new`
- **AND** `modelProviders` 数组中 SHALL NOT 出现两个 `deepseek` 配置项

#### Scenario: 覆盖时的其他字段

- **WHEN** 已有 `deepseek` 配置包含 `baseURL: "https://old.url"`
- **AND** 用户执行 `nfa models add deepseek --apiKey sk-new` 不指定 `--baseURL`
- **THEN** 系统 SHALL 将 `baseURL` 重置为空（使用默认值）

### Requirement: 配置持久化

系统 SHALL 将修改后的配置立即写入 `nfa.json` 文件。

#### Scenario: 配置写入成功

- **WHEN** 用户执行 `nfa models add deepseek --apiKey sk-xxx`
- **AND** 写入操作成功
- **THEN** 系统 SHALL 输出成功提示信息
- **AND** 系统 SHALL 将更新后的完整配置保存到 `nfa.json`

#### Scenario: 配置写入失败

- **WHEN** 写入配置文件时发生 IO 错误
- **THEN** 系统 SHALL 返回错误信息，包含具体的错误原因
