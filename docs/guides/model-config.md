# 模型配置 (Model Configuration)

NFA 支持多种模型提供商，通过灵活的配置实现不同场景下的最佳模型选择。

## 概述

模型配置是 NFA 的核心配置部分，决定了 Agent 的能力和行为。NFA 支持三种模型角色：

- **Main（主模型）** - 用于主要的对话、复杂推理和深度分析
- **Fast（快速模型）** - 用于简单事务、快速响应和轻量级任务
- **Vision（视觉模型）** - 用于图片理解、OCR 和视觉分析

通过合理配置不同的模型提供商和模型角色，可以在性能、成本和质量之间找到最佳平衡。

## 模型角色说明

### Main 模型

主模型是 Agent 的大脑，负责处理最重要的任务：

- 主要对话交互
- 复杂问题分析
- 金融数据深度解读
- 投资策略建议

**推荐模型**：选择能力最强、理解力最好的模型

### Fast 模型

快速模型用于处理简单和重复性任务：

- 话题分类
- 简单信息提取
- 快速查询响应
- 辅助性分析

**推荐模型**：选择响应速度快、成本较低的模型。如果未配置，自动回退到主模型。

### Vision 模型

视觉模型用于理解图像内容：

- 网页截图分析
- 图表解读
- OCR 文字识别
- 视觉问答

**推荐模型**：选择支持图像输入的模型（如 GPT-4 Vision）。如果未配置，部分视觉功能将不可用。

## 支持的模型提供商

### Ollama

Ollama 是本地模型运行平台，适合需要隐私保护或离线使用的场景。

> **BREAKING CHANGE**: 配置格式已更新！请参考下方的迁移指南。

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
    "light": "ollama/mistral",
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
        "cost": {"input": 0, "output": 0},
        "contextWindow": 4096,
        "maxOutputTokens": 2048
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

**模型配置字段 (ModelConfig)**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 模型名称 |
| `reasoning` | bool | 否 | 是否支持推理/思考模式（预留） |
| `vision` | bool | 否 | 是否支持视觉（预留） |
| `cost` | object | 否 | 价格信息，包含 `input` 和 `output` 字段（预留） |
| `contextWindow` | int | 否 | 上下文窗口大小（预留） |
| `maxOutputTokens` | int | 否 | 最大输出 Token 数（预留） |

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
          {"name": "deepseek-chat"},
          {"name": "deepseek-coder"},
          {"name": "deepseek-reasoner", "reasoning": true}
        ]
      }
    }
  ],
  "defaultModels": {
    "primary": "deepseek/deepseek-chat",
    "light": "deepseek/deepseek-chat",
    "vision": ""
  }
}
```

**配置参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `apiKey` | string | 是 | Deepseek API 密钥 |
| `models` | array | 否 | 模型配置列表，空列表表示不注册任何模型 |

**支持模型**:
- `deepseek-chat` - 通用对话模型
- `deepseek-coder` - 代码生成模型
- `deepseek-reasoner` - 思维链推理模型

**使用场景**:
- 需要高质量的中文对话能力
- 性价比优先
- 需要代码生成能力（使用 deepseek-coder）

### OpenAI Compatible

支持任何兼容 OpenAI API 的服务，如阿里云通义千问、智谱 AI 等。

**配置示例**:
```json
{
  "modelProviders": [
    {
      "openaiCompatible": {
        "name": "aliyun",
        "baseURL": "https://dashscope.aliyuncs.com/compatible-mode/v1",
        "apiKey": "your-aliyun-api-key",
        "models": [
          {
            "name": "qwen-max",
            "reasoning": true,
            "vision": false,
            "cost": {"input": 0.001, "output": 0.01},
            "contextWindow": 262144,
            "maxOutputTokens": 32768
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
    "primary": "aliyun/qwen-max",
    "light": "aliyun/qwen-turbo",
    "vision": "aliyun/qwen-vl-plus"
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
模型名称使用 `provider/model-name` 格式，如 `aliyun/qwen-max`。

**使用场景**:
- 使用其他云服务商的模型
- 需要自定义 API 端点
- 需要使用特定服务商的视觉模型

## 模型路由机制

NFA 根据任务类型自动选择合适的模型：

1. **主要对话和分析** → 使用 Main 模型
2. **话题分类等轻量任务** → 使用 Fast 模型
3. **图片相关任务** → 使用 Vision 模型

这种设计可以：
- 优化性能 - 简单任务用快速模型
- 降低成本 - 避免在大模型上处理简单请求
- 提升质量 - 复杂任务用最强模型

## 命令行参数覆盖

命令行参数可以临时覆盖配置文件中的模型设置：

```bash
# 使用指定的主模型
nfa --model "deepseek-chat" "分析一下当前市场"

# 使用指定的快速模型
nfa --light-model "ollama/mistral" "简单介绍一下ETF"

# 同时指定主模型和快速模型
nfa --model "deepseek-chat" --fast-model "ollama/mistral" "问题内容"
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
          {"name": "deepseek-chat"},
          {"name": "deepseek-reasoner", "reasoning": true}
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
    "primary": "deepseek/deepseek-chat",
    "light": "ollama/mistral",
    "vision": "openai/gpt-4-vision-preview"
  }
}
```

这个配置实现了：
- 主任务用 Deepseek（质量高）
- 快速任务用 Ollama Mistral（本地免费）
- 视觉任务用 GPT-4 Vision（能力最强）

## 迁移指南

### Ollama 配置迁移

**旧格式**（不再支持）:
```json
{
  "ollama": {
    "models": ["llama2", "mistral"]
  }
}
```

**新格式**:
```json
{
  "ollama": {
    "models": [
      {"name": "llama2"},
      {"name": "mistral"}
    ]
  }
}
```

### 迁移步骤

1. **备份现有配置**:
   ```bash
   cp ~/.nfa/nfa.json ~/.nfa/nfa.json.backup
   ```

2. **更新配置文件**:
   - 将 Ollama 的 `models` 字段从字符串数组改为对象数组
   - 为每个模型创建 `{"name": "模型名"}` 对象

3. **验证配置**:
   ```bash
   nfa models list
   ```

## 最佳实践

### 1. 根据场景选择模型

| 场景 | 推荐 | 原因 |
|------|------|------|
| 日常使用 | Ollama + Deepseek | 本地处理简单任务，复杂任务用 Deepseek |
| 需要视觉 | OpenAI Compatible | GPT-4 Vision 能力最强 |
| 纯代码工作 | Deepseek Coder | 专门优化的代码模型 |
| 隐私优先 | Ollama | 全本地运行 |

### 2. 设置合理的超时时间

对于 Ollama，根据硬件性能调整：
```json
{
  "ollama": {
    "timeout": 600  // 强大的硬件可以设置更长
  }
}
```

### 3. 配置模型元数据（预留字段）

虽然当前版本中 `reasoning`、`vision`、`cost` 等字段预留供日后使用，但建议在配置时声明这些信息，以便未来的智能路由功能使用：

```json
{
  "models": [
    {
      "name": "qwen-max",
      "reasoning": true,
      "vision": false,
      "cost": {"input": 0.001, "output": 0.01},
      "contextWindow": 262144
    }
  ]
}
```

### 4. 测试模型兼容性

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
