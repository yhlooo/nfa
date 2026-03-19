# 提案：添加 MiniMax 模型提供商支持

## 问题陈述

NFA 目前支持 Ollama、DeepSeek、Qwen、智谱 ZAI 等模型提供商，但不支持 MiniMax。用户希望使用 MiniMax 的 M2.5 和 M2.7 模型。

## 目标

添加 MiniMax 作为新的模型提供商，支持以下模型：
- `minimax-m2.5`
- `minimax-m2.7`

## 范围

### 包含

- 创建 `pkg/models/minimax.go` 实现 MiniMax 提供商
- 在 `pkg/models/providers.go` 中添加 MiniMax 配置字段
- 在 `pkg/agents/genkit.go` 中添加 MiniMax 提供商的注册逻辑
- 在 `pkg/models/i18n.go` 中添加模型描述消息
- 更新 `docs/reference/config.md` 文档

### 不包含

- MiniMax 特殊功能（如有）的深度集成
- 其他 MiniMax 模型的支持

## 方法

参考现有的 `zai.go` 和 `qwen.go` 实现模式：

1. MiniMax 提供 OpenAI 兼容 API，使用 `pkg/genkitplugins/oai` 插件
2. 预设推荐的模型列表（M2.5、M2.7）
3. 支持用户自定义模型配置

## 风险与未知

- MiniMax M2.5/M2.7 的具体参数（上下文窗口、价格等）需要后续验证
- MiniMax 是否支持 reasoning 模式及其 API 参数需要验证

## 成功标准

- 用户可以通过配置 `modelProviders[].minimax` 使用 MiniMax 模型
- 模型注册后可以在交互界面选择使用
