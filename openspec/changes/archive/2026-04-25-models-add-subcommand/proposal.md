## Why

当前 `nfa models` 命令仅有 `list` 子命令，用户只能查看模型列表但无法通过命令行添加模型供应商配置。用户必须手动编辑 `~/.nfa/nfa.json` 文件来配置新的模型供应商（如 DeepSeek、Qwen 等），操作繁琐、易出错。

## What Changes

- 为 `nfa models` 新增 `add` 子命令，支持通过命令行为已有供应商类型添加配置
- 支持以下供应商类型：`deepseek`、`qwen`、`moonshot`、`minimax`、`openrouter`、`z-ai`、`ollama`、`openai-compatible`
- 各供应商的必填和可选参数通过命令行 flag 指定（如 `--apiKey`、`--baseURL`、`--name`、`--serverAddress`、`--timeout`）
- 若供应商已配置则覆盖原有配置
- 新增的配置直接写入 `nfa.json` 配置文件，不做连接验证

## Capabilities

### New Capabilities

- `cli-model-add`: 通过 CLI 添加模型供应商配置

### Modified Capabilities

<!-- 无现有 capability 需要修改 -->

## Impact

- 影响文件：
  - `pkg/commands/models.go`：新增 `add` 子命令实现
  - `pkg/commands/i18n.go`：新增 i18n 消息定义
  - `pkg/configs/load.go`：可能需要新增按需保存配置的辅助函数
