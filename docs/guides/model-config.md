# 模型配置 (Model Configuration)

NFA 支持多种模型提供商，通过灵活的配置实现不同场景下的最佳模型选择。

## 概述

模型配置是 NFA 的核心配置部分，决定了 Agent 的能力和行为。NFA 支持两种模型角色：

- **Primary（主模型）** - 用于主要的对话、复杂推理和深度分析
- **Vision（视觉模型）** - 用于图片理解、OCR 和视觉分析

通过合理配置不同的模型提供商和模型角色，可以在性能、成本和质量之间找到最佳平衡。

## 模型角色说明

### Primary 模型

主模型是 Agent 的大脑，负责处理所有任务：

- 主要对话交互
- 复杂问题分析
- 金融数据深度解读
- 投资策略建议

**推荐模型**：选择能力最强、理解力最好的模型

### Vision 模型

视觉模型用于理解图像内容：

- 网页截图分析
- 图表解读
- OCR 文字识别
- 视觉问答

**推荐模型**：选择支持图像输入的模型（如 GPT-4 Vision、通义千问 VL）。如果未配置，自动回退到主模型。

## 支持的模型提供商

### Ollama

Ollama 是本地模型运行平台，适合需要隐私保护或离线使用的场景。

**配置示例**:
```json
{
  "modelProviders": [
    {
      "ollama": {
        "serverAddress": "http://localhost:11434",
        "timeout": 300,
        "models": [
          {"name": "llama2"},
          {"name": "mistral"},
          {"name": "codellama"}
        ]
      }
    }
  ],
  "defaultModels": {
    "primary": "ollama/llama2",
    "vision": ""
  }
}
```

**完整模型配置示例**（所有可选字段）:
```json
{
  "ollama": {
    "serverAddress": "http://localhost:11434",
    "timeout": 300,
    "models": [
      {
        "name": "llama2",
        "reasoning": false,
        "vision": false,
        "prices": {"input": 0, "output": 0},
        "contextWindow": 4096
      }
    ]
  }
}
```

**配置参数**:

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `serverAddress` | string | `http://localhost:11434` | Ollama 服务地址 |
| `timeout` | int | `300` | 请求超时时间（秒） |
| `models` | array | `[]` | 模型配置列表，空列表表示不注册任何模型 |

**模型配置字段**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 模型名称 |
| `reasoning` | bool | 否 | 是否支持推理/思考模式 |
| `vision` | bool | 否 | 是否支持视觉 |
| `prices` | object | 否 | 价格信息，包含 `input`、`output` 和 `cached` 字段（单位：每百万 Token） |
| `contextWindow` | int | 否 | 上下文窗口大小 |

**使用场景**:
- 需要本地运行，保护数据隐私
- 成本敏感，避免使用付费 API
- 需要离线使用

**注意事项**:
- 需要先安装和运行 Ollama
- 模型需要提前下载：`ollama pull llama2`
- 性能取决于本地硬件配置

### Deepseek

Deepseek 提供高性价比的中文大模型服务。

**配置示例**:
```json
{
  "modelProviders": [
    {
      "deepseek": {
        "apiKey": "your-deepseek-api-key",
        "models": [
          {"name": "deepseek-v4-pro"},
          {"name": "deepseek-v4-flash"}
        ]
      }
    }
  ],
  "defaultModels": {
    "primary": "deepseek/deepseek-v4-pro",
    "vision": ""
  }
}
```

**配置参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `apiKey` | string | 是 | Deepseek API 密钥 |
| `baseURL` | string | 否 | API 基础地址，默认 `https://api.deepseek.com` |
| `models` | array | 否 | 模型配置列表，空列表表示不注册任何模型 |

**预设模型**:
- `deepseek-v4-pro` - 推理能力强，1M 上下文
- `deepseek-v4-flash` - 轻量快速，1M 上下文

**使用场景**:
- 需要高质量的中文对话能力
- 性价比优先

### OpenAI Compatible

支持任何兼容 OpenAI API 的服务。

**配置示例**:
```json
{
  "modelProviders": [
    {
      "openaiCompatible": {
        "name": "qwen",
        "baseURL": "https://dashscope.aliyuncs.com/compatible-mode/v1",
        "apiKey": "your-qwen-api-key",
        "models": [
          {
            "name": "qwen-max",
            "reasoning": true,
            "vision": false,
            "prices": {"input": 0.001, "output": 0.01},
            "contextWindow": 262144
          },
          {
            "name": "qwen-turbo"
          },
          {
            "name": "qwen-vl-plus",
            "vision": true
          }
        ]
      }
    }
  ],
  "defaultModels": {
    "primary": "qwen/qwen-max",
    "vision": "qwen/qwen-vl-plus"
  }
}
```

**配置参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 提供商名称，用于模型命名 |
| `baseURL` | string | 是 | API 基础地址 |
| `apiKey` | string | 是 | API 密钥 |
| `models` | array | 否 | 模型配置列表，空列表表示不注册任何模型 |

**模型命名规则**:
模型名称使用 `provider/model-name` 格式，如 `qwen/qwen-max`。

**使用场景**:
- 使用其他云服务商的模型
- 需要自定义 API 端点
- 需要使用特定服务商的视觉模型

### ZAI（智谱 AI）

智谱 AI 提供 GLM 系列大模型服务。

**配置示例**:
```json
{
  "modelProviders": [
    {
      "z-ai": {
        "apiKey": "your-zai-api-key"
      }
    }
  ],
  "defaultModels": {
    "primary": "z-ai/glm-5.1",
    "vision": "z-ai/glm-5v-turbo"
  }
}
```

**配置参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `apiKey` | string | 是 | 智谱 AI API 密钥 |
| `baseURL` | string | 否 | API 基础地址，默认 `https://open.bigmodel.cn/api/paas/v4/` |
| `models` | array | 否 | 模型配置列表，默认使用预设模型 |

**预设模型**:
- `glm-5.1` - 支持推理，200K 上下文
- `glm-5` - 支持推理，200K 上下文
- `glm-5v-turbo` - 支持推理和视觉理解，200K 上下文

### Qwen（通义千问）

阿里云通义千问，中英文双语能力强。

**配置示例**:
```json
{
  "modelProviders": [
    {
      "qwen": {
        "apiKey": "your-qwen-api-key"
      }
    }
  ],
  "defaultModels": {
    "primary": "qwen/qwen3.6-plus",
    "vision": "qwen/qwen3.6-plus"
  }
}
```

**配置参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `apiKey` | string | 是 | 通义千问 API 密钥 |
| `baseURL` | string | 否 | API 基础地址，默认 `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| `models` | array | 否 | 模型配置列表，默认使用预设模型 |

**预设模型**:
- `qwen3.5-397b-a17b` - 支持推理和视觉理解，254K 上下文
- `qwen3.6-plus` - 支持推理和视觉理解，991K 上下文
- `qwen3.6-35b-a3b` - 支持推理和视觉理解，254K 上下文
- `qwen3.6-flash` - 支持推理和视觉理解，991K 上下文

### Moonshot AI（月之暗面）

月之暗面提供 Kimi 系列模型。

**配置示例**:
```json
{
  "modelProviders": [
    {
      "moonshotai": {
        "apiKey": "your-moonshot-api-key"
      }
    }
  ],
  "defaultModels": {
    "primary": "moonshotai/kimi-k2.6",
    "vision": "moonshotai/kimi-k2.6"
  }
}
```

**配置参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `apiKey` | string | 是 | 月之暗面 API 密钥 |
| `baseURL` | string | 否 | API 基础地址，默认 `https://api.moonshot.cn/v1` |
| `models` | array | 否 | 模型配置列表，默认使用预设模型 |

**预设模型**:
- `kimi-k2.6` - 支持推理和视觉理解，256K 上下文
- `kimi-k2.5` - 支持推理和视觉理解，256K 上下文

### MiniMax

MiniMax 提供大模型服务。

**配置示例**:
```json
{
  "modelProviders": [
    {
      "minimax": {
        "apiKey": "your-minimax-api-key"
      }
    }
  ],
  "defaultModels": {
    "primary": "minimax/minimax-m2.7",
    "vision": ""
  }
}
```

**配置参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `apiKey` | string | 是 | MiniMax API 密钥 |
| `baseURL` | string | 否 | API 基础地址，默认 `https://api.minimaxi.com/v1` |
| `models` | array | 否 | 模型配置列表，默认使用预设模型 |

**预设模型**:
- `minimax-m2.7` - 支持推理，200K 上下文

### OpenRouter

OpenRouter 提供统一的 API 接口访问多种模型。

**配置示例**:
```json
{
  "modelProviders": [
    {
      "openrouter": {
        "apiKey": "your-openrouter-api-key"
      }
    }
  ],
  "defaultModels": {
    "primary": "openrouter/openai/gpt-5.4",
    "vision": "openrouter/openai/gpt-5.4"
  }
}
```

**配置参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `apiKey` | string | 是 | OpenRouter API 密钥 |
| `baseURL` | string | 否 | API 基础地址，默认 `https://openrouter.ai/api/v1` |
| `models` | array | 否 | 模型配置列表，默认使用预设模型 |

**预设模型**:
- `google/gemini-3.1-pro-preview` - 支持推理和视觉理解，1M 上下文
- `openai/gpt-5.4` - 支持推理和视觉理解，1M 上下文
- `anthropic/claude-sonnet-4.6` - 支持推理和视觉理解，1M 上下文
- `anthropic/claude-opus-4.7` - 支持推理和视觉理解，1M 上下文
- `x-ai/grok-4.20` - 支持推理和视觉理解，2M 上下文

## 模型路由机制

NFA 根据任务类型自动选择合适的模型：

1. **主要对话和分析** → 使用 Primary 模型
2. **图片相关任务** → 使用 Vision 模型（如未配置则回退到 Primary 模型）

这种设计可以：
- 提升质量 - 为视觉任务选择专门的视觉模型
- 降低成本 - 避免为所有任务使用最昂贵的模型

## 命令行参数覆盖

命令行参数可以临时覆盖配置文件中的模型设置：

```bash
# 使用指定的主模型
nfa --model "deepseek-v4-pro" "分析一下当前市场"

# 使用指定的视觉模型
nfa --vision-model "qwen/qwen-vl-plus" "分析这张图表"

# 同时指定主模型和视觉模型
nfa --model "deepseek-v4-pro" --vision-model "qwen/qwen-vl-plus" "问题内容"
```

**优先级**：命令行参数 > 配置文件 > 默认值

这种设计非常适合临时测试不同模型，而不需要修改配置文件。

## 混合配置示例

可以同时配置多个模型提供商，为不同角色选择不同的提供商：

```json
{
  "modelProviders": [
    {
      "ollama": {
        "serverAddress": "http://localhost:11434",
        "timeout": 300,
        "models": [
          {"name": "mistral"},
          {"name": "llama2"}
        ]
      }
    },
    {
      "deepseek": {
        "apiKey": "your-deepseek-api-key",
        "models": [
          {"name": "deepseek-v4-pro"},
          {"name": "deepseek-v4-flash"}
        ]
      }
    },
    {
      "openaiCompatible": {
        "name": "openai",
        "baseURL": "https://api.openai.com/v1",
        "apiKey": "your-openai-api-key",
        "models": [
          {"name": "gpt-4-vision-preview", "vision": true}
        ]
      }
    }
  ],
  "defaultModels": {
    "primary": "deepseek/deepseek-v4-pro",
    "vision": "openai/gpt-4-vision-preview"
  }
}
```

这个配置实现了：
- 主任务用 Deepseek（质量高）
- 视觉任务用 GPT-4 Vision（能力最强）
- 本地可用 Ollama 模型作为备选

## 最佳实践

### 1. 根据场景选择模型

| 场景 | 推荐 | 原因 |
|------|------|------|
| 日常使用 | Ollama + Deepseek | 本地处理简单任务，复杂任务用 Deepseek |
| 需要视觉 | Qwen / OpenAI Compatible | 视觉理解能力强 |
| 隐私优先 | Ollama | 全本地运行 |
| 多模型切换 | OpenRouter | 一个 API 访问多种模型 |

### 2. 设置合理的超时时间

对于 Ollama，根据硬件性能调整：
```json
{
  "ollama": {
    "timeout": 600
  }
}
```

### 3. 测试模型兼容性

新模型上线前，先用命令行参数测试：
```bash
nfa --model "new-model-name" "测试一下"
```

## 故障排查

### 模型加载失败

1. **检查配置文件语法**
   ```bash
   # 验证 JSON 格式
   cat ~/.nfa/nfa.json | jq .
   ```

2. **检查 API 密钥**
   - 确认密钥格式正确
   - 确认密钥未过期

3. **检查网络连接**
   - 对于云端模型，确保可以访问 API

4. **查看日志**
   ```bash
   tail -f ~/.nfa/nfa.log
   ```

### 没有模型被注册

如果配置了 `models` 字段但仍然显示 "no models configured"：

1. **检查 models 数组格式**
   - 确保是对象数组而不是字符串数组
   - 每个对象必须包含 `name` 字段

2. **检查 provider 是否启用**
   - 确认配置在 `modelProviders` 数组中
   - 确认 provider 名称拼写正确

### Ollama 连接失败

1. **确认 Ollama 正在运行**
   ```bash
   ollama list
   ```

2. **检查服务地址**
   ```bash
   curl http://localhost:11434/api/tags
   ```

3. **确认模型已下载**
   ```bash
   ollama pull llama2
   ```

### 视觉模型不工作

1. **确认 vision 字段已配置**
   ```json
   {
     "defaultModels": {
       "vision": "openai/gpt-4-vision-preview"
     }
   }
   ```

2. **确认模型支持视觉**
   - 查看模型文档确认能力

3. **测试 WebBrowse 工具**
   - 让 Agent 尝试读取一个网页图片

## 查看可用模型

NFA 提供命令查看当前配置的可用模型：

```bash
# 列出所有可用模型
nfa models list
```

这可以帮助你确认模型配置是否正确。
