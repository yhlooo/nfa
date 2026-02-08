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

### defaultModels

默认模型配置，指定不同任务类型使用的模型：

```json
{
  "defaultModels": {
    "main": "ollama/llama2",
    "fast": "ollama/llama2",
    "vision": ""
  }
}
```

字段说明：
- `main` - 主模型，用于回答用户问题
- `fast` - 快速模型，用于处理简单事务。为空时使用主模型
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
    "main": "ollama/llama2",
    "fast": "ollama/llama2",
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
