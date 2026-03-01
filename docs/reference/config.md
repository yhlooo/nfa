# 配置参考

NFA 的配置文件位于 `~/.nfa/nfa.json`。

## 快速配置

在 `~/.nfa/nfa.json` 中添加配置。配置结构参考 [config.go](../../pkg/configs/config.go)

## 配置结构

```json
{
  "modelProviders": [...],
  "defaultModels": {...},
  "dataProviders": [...]
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
        "models": ["llama2"]
      }
    }
  ]
}
```

字段说明：
- `serverAddress` - Ollama 服务端地址，默认 `http://localhost:11434`
- `timeout` - 模型响应超时时间（秒），默认 `300`
- `models` - 模型名列表（数组），空表示使用 Ollama 已下载的所有模型

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

**模型描述字段**:

所有模型提供商都支持为每个模型添加描述信息，帮助你在交互式选择菜单中了解模型特点。

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
            "description": "Meta 的开源模型，擅长通用对话和推理"
          },
          {
            "name": "qwen3:14b",
            "description": "阿里云通义千问，中英文双语能力强，适合金融分析"
          },
          {
            "name": "mistral",
            "description": "轻量级快速模型，适合简单任务"
          }
        ]
      }
    }
  ]
}
```

**模型配置字段**:

每个模型可以配置以下字段：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | ✓ | 模型名称 |
| `description` | string | - | 模型描述，显示在选择菜单中 |
| `reasoning` | boolean | - | 是否支持推理/思考模式 |
| `vision` | boolean | - | 是否支持视觉/图片理解 |
| `cost.input` | float | - | 每 1K 输入 Token 价格 |
| `cost.output` | float | - | 每 1K 输出 Token 价格 |
| `cost.cached` | float | - | 每 1K 缓存 Token 价格 |
| `contextWindow` | int64 | - | 上下文窗口大小（Token 数） |
| `maxOutputTokens` | int64 | - | 最大输出 Token 数 |

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
            "description": "Meta Llama 3.2，通用对话能力强，适合复杂推理任务",
            "reasoning": true,
            "vision": false,
            "contextWindow": 128000,
            "maxOutputTokens": 4096
          },
          {
            "name": "qwen3:14b",
            "description": "通义千问 3 14B，中英双语平衡，金融领域表现优秀",
            "reasoning": true,
            "vision": false,
            "contextWindow": 32768,
            "maxOutputTokens": 8192
          },
          {
            "name": "mistral",
            "description": "Mistral 7B，轻量快速，适合简单问答和快速响应",
            "reasoning": false,
            "vision": false,
            "contextWindow": 8192,
            "maxOutputTokens": 2048
          }
        ]
      }
    },
    {
      "deepseek": {
        "apiKey": "your-api-key",
        "models": [
          {
            "name": "deepseek-chat",
            "description": "DeepSeek Chat，深度推理能力强，适合复杂分析",
            "reasoning": true,
            "vision": false,
            "cost": {
              "input": 0.14,
              "output": 0.28
            }
          }
        ]
      }
    }
  ],
  "defaultModels": {
    "primary": "ollama/llama3.2",
    "light": "ollama/mistral",
    "vision": ""
  }
}
```

**描述显示效果**:

在交互式模型选择菜单中，描述会显示在模型名称后面：

```
Select primary model

 ❯ 1. ollama/llama3.2 - Meta Llama 3.2，通用对话能力强，适合复杂推理任务
   2. ollama/qwen3:14b - 通义千问 3 14B，中英双语平衡，金融领域表现优秀
   3. ollama/mistral - Mistral 7B，轻量快速，适合简单问答和快速响应
   4. deepseek/deepseek-chat - DeepSeek Chat，深度推理能力强，适合复杂分析
```

描述超过 80 字符时会自动截断并显示 "..."。

### defaultModels

默认模型配置，指定不同任务类型使用的模型：

```json
{
  "defaultModels": {
    "primary": "ollama/llama2",
    "light": "ollama/llama2",
    "vision": ""
  }
}
```

字段说明：
- `primary` - 主模型，用于回答用户问题
- `light` - 轻量模型，用于处理简单事务。为空时使用主模型
- `vision` - 视觉模型，用于处理图片理解任务

### dataProviders

数据提供商配置数组。支持以下提供商：

#### Alpha Vantage

```json
{
  "dataProviders": [
    {
      "alphaVantage": {
        "apiKey": "your-api-key"
      }
    }
  ]
}
```

#### Tencent Cloud Web Search Agent

```json
{
  "dataProviders": [
    {
      "tcloudWSA": {
        "secretID": "your-secret-id",
        "secretKey": "your-secret-key",
        "endpoint": ""
      }
    }
  ]
}
```

字段说明：
- `secretID` - 腾讯云 Secret ID
- `secretKey` - 腾讯云 Secret Key
- `endpoint` - 服务端点（可选）

## 完整配置示例

```json
{
  "modelProviders": [
    {
      "ollama": {
        "serverAddress": "http://localhost:11434",
        "timeout": 300,
        "models": ["llama2"]
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
    "light": "ollama/llama2",
    "vision": ""
  },
  "dataProviders": [
    {
      "alphaVantage": {
        "apiKey": "your-alpha-vantage-api-key"
      }
    },
    {
      "tcloudWSA": {
        "secretID": "your-secret-id",
        "secretKey": "your-secret-key",
        "endpoint": ""
      }
    }
  ]
}
```
