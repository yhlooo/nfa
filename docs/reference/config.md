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
        "baseURL": "http://localhost:11434",
        "model": "llama2"
      }
    }
  ]
}
```

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
        "baseURL": "https://api.example.com/v1",
        "apiKey": "your-api-key"
      }
    }
  ]
}
```

### defaultModels

默认模型配置，指定不同任务类型使用的模型：

```json
{
  "defaultModels": {
    "chat": "llama2",
    "reasoning": "llama2"
  }
}
```

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
        "secretId": "your-secret-id",
        "secretKey": "your-secret-key",
        "region": "ap-guangzhou"
      }
    }
  ]
}
```

## 完整配置示例

```json
{
  "modelProviders": [
    {
      "ollama": {
        "baseURL": "http://localhost:11434",
        "model": "llama2"
      }
    },
    {
      "deepseek": {
        "apiKey": "your-api-key"
      }
    }
  ],
  "defaultModels": {
    "chat": "llama2",
    "reasoning": "llama2"
  },
  "dataProviders": [
    {
      "alphaVantage": {
        "apiKey": "your-alpha-vantage-api-key"
      }
    },
    {
      "tcloudWSA": {
        "secretId": "your-secret-id",
        "secretKey": "your-secret-key",
        "region": "ap-guangzhou"
      }
    }
  ]
}
```
