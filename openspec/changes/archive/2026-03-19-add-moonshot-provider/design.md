# 设计文档：Moonshot AI 提供商

## 架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    配置层 (JSON)                                 │
├─────────────────────────────────────────────────────────────────┤
│  modelProviders:                                                │
│    - moonshot:                                                  │
│        apiKey: "xxx"                                            │
│        baseURL: "https://api.moonshot.cn/v1"  # 可选            │
│        models: [...]                          # 可选            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    类型定义层                                    │
├─────────────────────────────────────────────────────────────────┤
│  providers.go                                                   │
│    ModelProvider.Moonshot *MoonshotOptions                      │
│                                                                 │
│  moonshot.go                                                    │
│    MoonshotOptions { BaseURL, APIKey, Models }                  │
│    MoonshotModels(ctx) []ModelConfig                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    插件层                                        │
├─────────────────────────────────────────────────────────────────┤
│  genkitplugins/oai/openai_compatible.go                         │
│    OpenAICompatible 插件 (复用现有实现)                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    注册层                                        │
├─────────────────────────────────────────────────────────────────┤
│  agents/genkit.go                                               │
│    case p.Moonshot != nil:                                      │
│      plugin := p.Moonshot.Plugin()                              │
│      p.Moonshot.RegisterModels(ctx, g, plugin)                  │
└─────────────────────────────────────────────────────────────────┘
```

## 文件变更

### 新建文件

| 文件 | 描述 |
|------|------|
| `pkg/models/moonshot.go` | Moonshot 提供商实现 |

### 修改文件

| 文件 | 变更内容 |
|------|----------|
| `pkg/models/providers.go` | 添加 `Moonshot *MoonshotOptions` 字段 |
| `pkg/models/i18n.go` | 添加 Kimi 模型描述消息 |
| `pkg/i18n/active.zh.yaml` | 添加中文翻译 |
| `pkg/i18n/active.en.yaml` | 添加英文翻译 |
| `pkg/agents/genkit.go` | 添加 Moonshot 提供商的注册逻辑 |
| `docs/reference/config.md` | 添加 Moonshot 配置说明 |

## Kimi K2.5 模型配置

```go
ModelConfig{
    Name:        "kimi-k2.5",
    Description: "月之暗面 Kimi K2.5，支持推理与视觉理解",
    Reasoning:   true,
    Vision:      true,
    Cost: ModelCost{
        Input:  0.004,  // 4/M tokens = 0.004/1K tokens
        Output: 0.021,  // 21/M tokens = 0.021/1K tokens
    },
    ContextWindow:   256000,
    MaxOutputTokens: 256000,
}
```

## 配置示例

```json
{
  "modelProviders": [
    {
      "moonshotai": {
        "apiKey": "your-moonshot-api-key"
      }
    }
  ]
}
```

## 命名约定

- 提供商名称：`moonshotai`
- 模型完整名称：`moonshotai/kimi-k2.5`
- JSON 配置键：`moonshotai`
