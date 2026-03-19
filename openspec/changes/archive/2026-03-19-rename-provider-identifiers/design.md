## 设计决策

### D1: JSON 键名使用品牌名

**决策**：配置文件中的供应商键名使用品牌名而非公司名。

**理由**：
- 用户更熟悉品牌名（Qwen、智谱 AI）而非公司名（阿里云、智谱）
- 与其他供应商保持一致（ollama、deepseek 都是品牌名）

**示例**：
```json
{
  "modelProviders": [
    { "qwen": { "apiKey": "...", "models": [...] } },
    { "z-ai": { "apiKey": "...", "models": [...] } }
  ]
}
```

### D2: Go 标识符命名

**决策**：Go 代码中使用 Pascal Case 命名。

| JSON 键 | Go 字段名 | Go 类型名 | Go 常量名 |
|---------|----------|----------|----------|
| `qwen` | `Qwen` | `QwenOptions` | `QwenProviderName` |
| `z-ai` | `ZAI` | `ZAIOptions` | `ZAIProviderName` |

**理由**：
- `ZAI` 作为缩写全大写，符合 Go 命名惯例（类似 `URL`、`HTTP`）
- 与 `Qwen` 三字母对称，视觉一致

### D3: 文件命名

**决策**：使用简洁的品牌名作为文件名。

| 原文件 | 新文件 |
|--------|--------|
| `aliyun_dashscope.go` | `qwen.go` |
| `zhipu_bigmodel.go` | `zai.go` |

**理由**：
- 原文件名中的 `dashscope`、`bigmodel` 是 API 端点名称，属于实现细节
- 品牌名作为文件名更简洁直观

### D4: 不提供向后兼容

**决策**：直接使用新名称，不支持旧配置格式。

**理由**：
- 项目处于早期阶段，用户基数小
- 保持代码简洁，避免兼容层增加复杂度
- 用户需手动更新配置文件

## 实现方案

### 代码变更

#### 1. `pkg/models/providers.go`

```go
// 修改前
type ModelProvider struct {
    Ollama           *OllamaOptions           `json:"ollama,omitempty"`
    Zhipu            *BigModelOptions         `json:"zhipu,omitempty"`
    Aliyun           *DashScopeOptions        `json:"aliyun,omitempty"`
    Deepseek         *DeepseekOptions         `json:"deepseek,omitempty"`
    OpenAICompatible *OpenAICompatibleOptions `json:"openaiCompatible,omitempty"`
}

// 修改后
type ModelProvider struct {
    Ollama           *OllamaOptions           `json:"ollama,omitempty"`
    ZAI              *ZAIOptions              `json:"z-ai,omitempty"`
    Qwen             *QwenOptions             `json:"qwen,omitempty"`
    Deepseek         *DeepseekOptions         `json:"deepseek,omitempty"`
    OpenAICompatible *OpenAICompatibleOptions `json:"openaiCompatible,omitempty"`
}
```

#### 2. `pkg/models/qwen.go` (原 `aliyun_dashscope.go`)

```go
// 修改前
const (
    DashScopeProviderName = "aliyun"
    DashScopeBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

func DashScopeModels(ctx context.Context) []ModelConfig { ... }
type DashScopeOptions struct { ... }

// 修改后
const (
    QwenProviderName = "qwen"
    QwenBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

func QwenModels(ctx context.Context) []ModelConfig { ... }
type QwenOptions struct { ... }
```

#### 3. `pkg/models/zai.go` (原 `zhipu_bigmodel.go`)

```go
// 修改前
const (
    BigModelProviderName = "zhipu"
    BigModelBaseURL = "https://open.bigmodel.cn/api/paas/v4/"
)

func BigModelModels(ctx context.Context) []ModelConfig { ... }
type BigModelOptions struct { ... }

// 修改后
const (
    ZAIProviderName = "z-ai"
    ZAIBaseURL = "https://open.bigmodel.cn/api/paas/v4/"
)

func ZAIModels(ctx context.Context) []ModelConfig { ... }
type ZAIOptions struct { ... }
```

#### 4. `pkg/agents/genkit.go`

更新所有字段引用和错误日志消息。

### 文件操作

```bash
# 重命名文件
git mv pkg/models/aliyun_dashscope.go pkg/models/qwen.go
git mv pkg/models/zhipu_bigmodel.go pkg/models/zai.go
```

### 文档更新

更新以下文档中的所有示例：
- `docs/guides/model-config.md`
- `docs/guides/command-line.md`

### Spec 更新

更新活跃的 spec 文件：
- `openspec/specs/model-config/spec.md`
- `openspec/specs/model-selection/spec.md`
