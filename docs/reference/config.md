# 配置参考

NFA 的配置文件位于 `~/.nfa/nfa.json`。

## 快速配置

在 `~/.nfa/nfa.json` 中添加配置。配置结构参考 [config.go](../../pkg/configs/config.go)

## 配置结构

```json
{
  "modelProviders": [...],
  "defaultModels": {...},
  "dataProviders": {...},
  "channels": {...},
  "language": "zh",
  "maxContextWindow": 200000
}
```

## 配置项

### modelProviders

模型提供商配置数组。支持以下提供商：

#### Ollama

```json
{
  "modelProviders": [
    {
      "ollama": {
        "serverAddress": "http://localhost:11434",
        "timeout": 300,
        "models": [
          {"name": "llama2"}
        ]
      }
    }
  ]
}
```

字段说明：
- `serverAddress` - Ollama 服务端地址，默认 `http://localhost:11434`
- `timeout` - 模型响应超时时间（秒），默认 `300`
- `models` - 模型配置列表（对象数组），空表示使用 Ollama 已下载的所有模型

#### Deepseek

```json
{
  "modelProviders": [
    {
      "deepseek": {
        "apiKey": "your-api-key"
      }
    }
  ]
}
```

字段说明：
- `apiKey` - API 密钥（必填）
- `baseURL` - API 基础地址（可选，默认 `https://api.deepseek.com`）
- `models` - 模型配置列表（可选，默认使用预设模型）

预设模型：
- `deepseek-v4-pro` - 支持推理，1M 上下文
- `deepseek-v4-flash` - 支持推理，1M 上下文，轻量快速

#### OpenAI Compatible

```json
{
  "modelProviders": [
    {
      "openaiCompatible": {
        "name": "custom",
        "baseURL": "https://api.example.com/v1",
        "apiKey": "your-api-key"
      }
    }
  ]
}
```

字段说明：
- `name` - 提供商名称，用于标识模型来源
- `baseURL` - API 基础地址
- `apiKey` - API 密钥

#### ZAI（智谱 AI）

```json
{
  "modelProviders": [
    {
      "z-ai": {
        "apiKey": "your-api-key"
      }
    }
  ]
}
```

字段说明：
- `apiKey` - API 密钥（必填）
- `baseURL` - API 基础地址（可选，默认 `https://open.bigmodel.cn/api/paas/v4/`）
- `models` - 模型配置列表（可选，默认使用预设模型）

预设模型：
- `glm-5.1` - 支持推理，200K 上下文
- `glm-5` - 支持推理，200K 上下文
- `glm-5v-turbo` - 支持推理和视觉理解，200K 上下文

#### Qwen（通义千问）

```json
{
  "modelProviders": [
    {
      "qwen": {
        "apiKey": "your-api-key"
      }
    }
  ]
}
```

字段说明：
- `apiKey` - API 密钥（必填）
- `baseURL` - API 基础地址（可选，默认 `https://dashscope.aliyuncs.com/compatible-mode/v1`）
- `models` - 模型配置列表（可选，默认使用预设模型）

预设模型：
- `qwen3.5-397b-a17b` - 支持推理和视觉理解，254K 上下文
- `qwen3.6-plus` - 支持推理和视觉理解，991K 上下文
- `qwen3.6-35b-a3b` - 支持推理和视觉理解，254K 上下文
- `qwen3.6-flash` - 支持推理和视觉理解，991K 上下文

#### Moonshot AI（月之暗面）

```json
{
  "modelProviders": [
    {
      "moonshotai": {
        "apiKey": "your-api-key"
      }
    }
  ]
}
```

字段说明：
- `apiKey` - API 密钥（必填）
- `baseURL` - API 基础地址（可选，默认 `https://api.moonshot.cn/v1`）
- `models` - 模型配置列表（可选，默认使用预设模型）

预设模型：
- `kimi-k2.6` - 支持推理和视觉理解，256K 上下文
- `kimi-k2.5` - 支持推理和视觉理解，256K 上下文

#### MiniMax

```json
{
  "modelProviders": [
    {
      "minimax": {
        "apiKey": "your-api-key"
      }
    }
  ]
}
```

字段说明：
- `apiKey` - API 密钥（必填）
- `baseURL` - API 基础地址（可选，默认 `https://api.minimaxi.com/v1`）
- `models` - 模型配置列表（可选，默认使用预设模型）

预设模型：
- `minimax-m2.7` - 支持推理，200K 上下文

#### OpenRouter

```json
{
  "modelProviders": [
    {
      "openrouter": {
        "apiKey": "your-api-key"
      }
    }
  ]
}
```

字段说明：
- `apiKey` - API 密钥（必填）
- `baseURL` - API 基础地址（可选，默认 `https://openrouter.ai/api/v1`）
- `models` - 模型配置列表（可选，默认使用预设模型）

预设模型：
- `google/gemini-3.1-pro-preview` - 支持推理和视觉理解，1M 上下文
- `openai/gpt-5.4` - 支持推理和视觉理解，1M 上下文
- `anthropic/claude-sonnet-4.6` - 支持推理和视觉理解，1M 上下文
- `anthropic/claude-opus-4.7` - 支持推理和视觉理解，1M 上下文
- `x-ai/grok-4.20` - 支持推理和视觉理解，2M 上下文

**模型配置字段**:

每个模型可以配置以下字段：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 模型名称 |
| `reasoning` | boolean | 否 | 是否支持推理/思考模式 |
| `vision` | boolean | 否 | 是否支持视觉/图片理解 |
| `prices.input` | float | 否 | 每百万输入 Token 价格 |
| `prices.output` | float | 否 | 每百万输出 Token 价格 |
| `prices.cached` | float | 否 | 每百万缓存 Token 价格 |
| `contextWindow` | int64 | 否 | 上下文窗口大小（Token 数） |

**完整配置示例**:

```json
{
  "modelProviders": [
    {
      "ollama": {
        "serverAddress": "http://localhost:11434",
        "timeout": 300,
        "models": [
          {
            "name": "llama3.2",
            "reasoning": true,
            "vision": false,
            "contextWindow": 128000
          },
          {
            "name": "qwen3:14b",
            "reasoning": true,
            "vision": false,
            "contextWindow": 32768
          },
          {
            "name": "mistral",
            "reasoning": false,
            "vision": false,
            "contextWindow": 8192
          }
        ]
      }
    },
    {
      "deepseek": {
        "apiKey": "your-api-key",
        "models": [
          {
            "name": "deepseek-v4-pro",
            "reasoning": true,
            "vision": false,
            "prices": {
              "input": 12,
              "output": 24
            }
          }
        ]
      }
    }
  ],
  "defaultModels": {
    "primary": "ollama/llama3.2",
    "vision": ""
  }
}
```

### defaultModels

默认模型配置，指定不同任务类型使用的模型：

```json
{
  "defaultModels": {
    "primary": "ollama/llama2",
    "vision": ""
  }
}
```

字段说明：
- `primary` - 主模型，用于回答用户问题
- `vision` - 视觉模型，用于处理图片理解任务。为空时自动回退到主模型

### dataProviders

数据提供商配置对象。每种数据提供商最多配置一个实例。支持以下提供商：

#### Alpha Vantage

```json
{
  "dataProviders": {
    "alphaVantage": {
      "apiKey": "your-api-key"
    }
  }
}
```

#### Tencent Cloud Web Search Agent

```json
{
  "dataProviders": {
    "tcloudWSA": {
      "secretID": "your-secret-id",
      "secretKey": "your-secret-key",
      "endpoint": ""
    }
  }
}
```

字段说明：
- `secretID` - 腾讯云 Secret ID
- `secretKey` - 腾讯云 Secret Key
- `endpoint` - 服务端点（可选）

### channels

消息通道配置，用于通过外部平台与 Agent 交互。

```json
{
  "channels": {
    "enabled": true,
    "channels": [
      {
        "wecomAIBot": {
          "botID": "your-bot-id",
          "secret": "your-bot-secret"
        }
      },
      {
        "yuanbaoBot": {
          "appID": "your-app-id",
          "appSecret": "your-app-secret"
        }
      }
    ]
  }
}
```

字段说明：
- `enabled` - 是否启用消息通道
- `channels` - 通道配置列表

#### 企业微信智能机器人

- `botID` - 机器人 ID（必填）
- `secret` - 机器人密钥（必填）
- `url` - 自定义回调 URL（可选）

#### 元宝机器人

- `appID` - 应用 ID（必填）
- `appSecret` - 应用密钥（必填）
- `baseURL` - API 基础地址（可选）
- `websocketURL` - WebSocket 地址（可选）

### language

设置界面语言，可选值为 `en`（英文）或 `zh`（中文）。不设置时自动检测系统语言。

```json
{
  "language": "zh"
}
```

### maxContextWindow

最大上下文窗口大小（Token 数），默认 200K。用于限制 Agent 对话的上下文长度。

```json
{
  "maxContextWindow": 200000
}
```

## 完整配置示例

```json
{
  "modelProviders": [
    {
      "ollama": {
        "serverAddress": "http://localhost:11434",
        "timeout": 300,
        "models": [{"name": "llama2"}]
      }
    },
    {
      "deepseek": {
        "apiKey": "your-api-key"
      }
    },
    {
      "openaiCompatible": {
        "name": "custom",
        "baseURL": "https://api.example.com/v1",
        "apiKey": "your-api-key"
      }
    }
  ],
  "defaultModels": {
    "primary": "ollama/llama2",
    "vision": ""
  },
  "dataProviders": {
    "alphaVantage": {
      "apiKey": "your-alpha-vantage-api-key"
    },
    "tcloudWSA": {
      "secretID": "your-secret-id",
      "secretKey": "your-secret-key",
      "endpoint": ""
    }
  },
  "channels": {
    "enabled": false
  },
  "language": "zh"
}
```
