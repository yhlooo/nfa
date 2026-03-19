# 设计：MiniMax 提供商实现

## 架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                      MiniMax 提供商架构                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   用户配置 ~/.nfa/nfa.json                                          │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ {                                                           │   │
│   │   "modelProviders": [{                                      │   │
│   │     "minimax": {                                            │   │
│   │       "apiKey": "xxx",                                      │   │
│   │       "models": [...]  // 可选                              │   │
│   │     }                                                       │   │
│   │   }]                                                        │   │
│   │ }                                                           │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                              ▼                                      │
│   pkg/models/minimax.go                                             │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ MinimaxOptions {                                            │   │
│   │   BaseURL  // 默认 https://api.minimax.chat/v1              │   │
│   │   APIKey                                                     │   │
│   │   Models []ModelConfig                                       │   │
│   │ }                                                            │   │
│   │                                                              │   │
│   │ MinimaxModels() → 预设模型列表                               │   │
│   │ Plugin() → *oai.OpenAICompatible                             │   │
│   │ RegisterModels() → 注册模型到 Genkit                         │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                              ▼                                      │
│   pkg/genkitplugins/oai/openai_compatible.go                        │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │ OpenAI 兼容插件，处理 API 调用                               │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## 实现细节

### 1. MinimaxOptions 结构体

```go
type MinimaxOptions struct {
    BaseURL string        `json:"baseURL,omitempty"`
    APIKey  string        `json:"apiKey"`
    Models  []ModelConfig `json:"models,omitempty"`
}
```

### 2. 预设模型

| 模型 | Reasoning | Vision | 上下文窗口 | 最大输出 | 输入价格 | 输出价格 |
|------|-----------|--------|-----------|---------|---------|---------|
| minimax-m2.5 | true | false | 128000 | 8192 | 0.002 | 0.006 |
| minimax-m2.7 | true | false | 256000 | 16384 | 0.003 | 0.009 |

> 注：以上参数为估计值，需要后续验证修正

### 3. Reasoning 支持

参考 ZAI 的实现，使用 `OpenAICompatibleOptions.RegisterModels` 方法，该方式使用：
- `thinking.type: enabled` 开启 reasoning
- `thinking.type: disabled` 关闭 reasoning
- `reasoning_content` 字段获取思考内容

### 4. 配置示例

```json
{
  "modelProviders": [
    {
      "minimax": {
        "apiKey": "your-api-key",
        "models": [
          {
            "name": "minimax-m2.5",
            "description": "MiniMax M2.5 模型"
          }
        ]
      }
    }
  ]
}
```

## 文件变更清单

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `pkg/models/minimax.go` | 新增 | MiniMax 提供商实现 |
| `pkg/models/providers.go` | 修改 | 添加 `Minimax *MinimaxOptions` 字段 |
| `pkg/models/i18n.go` | 修改 | 添加模型描述消息 |
| `pkg/agents/genkit.go` | 修改 | 添加 MiniMax 提供商注册逻辑 |
| `docs/reference/config.md` | 修改 | 添加 MiniMax 配置说明 |
