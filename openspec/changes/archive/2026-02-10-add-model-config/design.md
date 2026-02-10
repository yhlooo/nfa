## Context

### Current State

当前系统中，模型供应商通过 HTTP API 动态发现可用模型：

- **Ollama**: 调用 `GET /api/tags` 获取已下载的模型列表
- **Deepseek**: 调用 OpenAI SDK 的 `Models.List()` 方法
- **OpenAI Compatible**: 调用 `GET /models` 端点

这种设计的问题：
1. 启动时需要阻塞等待 API 响应
2. API 返回仅包含模型名称，不包含能力信息
3. 模型能力（thinking、vision）需要硬编码判断

### Constraints

- 必须保持 Genkit 框架的 `ModelSupports` 结构兼容
- 不能破坏现有的模型命名格式 (`provider/model-name`)
- 配置解析使用标准库 `encoding/json`

## Goals / Non-Goals

**Goals:**
- 支持通过配置文件声明模型及其元数据
- 消除启动时的模型发现 API 调用
- 为未来的智能路由和成本计算提供数据基础
- 保持配置 JSON 格式简洁易读

**Non-Goals:**
- 本次变更不实现基于元数据的模型路由逻辑
- 不实现成本计算功能
- 不实现 token 计数和上下文管理
- 不提供配置迁移工具（用户需手动更新）

## Decisions

### 1. ModelConfig 结构定义

```go
// ModelConfig 模型配置
type ModelConfig struct {
    // 模型名称（必需）
    Name string `json:"name"`

    // 是否支持推理/思考模式（预留）
    Reasoning bool `json:"reasoning,omitempty"`

    // 是否支持视觉/图片理解（预留）
    Vision bool `json:"vision,omitempty"`

    // 价格信息（预留），单位：元/千Token
    Cost ModelCost `json:"cost,omitempty"`

    // 上下文窗口大小（预留）
    ContextWindow int `json:"contextWindow,omitempty"`

    // 最大输出 Token 数（预留）
    MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
}

// ModelCost 价格信息
type ModelCost struct {
    Input  float64 `json:"input,omitempty"`
    Output float64 `json:"output,omitempty"`
}
```

**Rationale:**
- `Name` 作为唯一必需字段，降低配置门槛
- 元数据字段使用 `omitempty` 标签，未配置时为 Go 零值
- `Cost` 使用结构体而非指针，简化判空逻辑（零值即免费或未知）
- `ModelSupports` 中已有的能力（Multiturn、Tools、SystemRole）由 Genkit 插件自动处理，无需重复配置

### 2. Provider Options 更新

各 Provider 统一添加 `Models []ModelConfig` 字段：

```go
// OllamaOptions
type OllamaOptions struct {
    ServerAddress string        `json:"serverAddress,omitempty"`
    Timeout       int           `json:"timeout,omitempty"`
    Models        []ModelConfig `json:"models,omitempty"`  // 从 []string 改变
}

// DeepseekOptions
type DeepseekOptions struct {
    APIKey string        `json:"apiKey"`
    Models []ModelConfig `json:"models,omitempty"`  // 新增
}

// OpenAICompatibleOptions
type OpenAICompatibleOptions struct {
    Name    string        `json:"name"`
    BaseURL string        `json:"baseURL"`
    APIKey  string        `json:"apiKey"`
    Models  []ModelConfig `json:"models,omitempty"`  // 新增
}
```

**Rationale:**
- 统一的数据结构，便于代码复用
- 明确的语义：配置了 models 就只使用配置的模型

### 3. 模型注册逻辑变更

#### 当前流程（API 发现）:
```
InitGenkit → NewGenkitWithModels → 调用 Provider API 列出模型 → 注册模型
```

#### 新流程（配置驱动）:
```
InitGenkit → NewGenkitWithModels → 遍历配置的 models → 注册模型
```

**关键变更:**

1. **移除 API 调用**:
   - 删除 `OllamaOptions.ListModels()` 调用
   - 删除 `models.ListOpenAICompatibleModels()` 调用
   - 删除 `deepseekPlugin.RegisterModels()` 中的 `client.Models.List()` 调用

2. **空配置处理**:
   - `models` 为空或未配置 → 不注册任何模型
   - 不再回退到 API 发现

3. **模型注册**:
   - 使用 `ModelConfig.Name` 作为模型名
   - 暂时忽略其他元数据字段（存储但不使用）

### 4. 代码组织

```
pkg/models/
├── providers.go       // 新增：ModelConfig 和 ModelCost 定义
├── ollama.go          // 修改：OllamaOptions.Models 类型
├── deepseek.go        // 修改：DeepseekOptions 新增 Models
└── openai_compatible.go  // 修改：OpenAICompatibleOptions 新增 Models

pkg/agents/
└── genkit.go          // 修改：NewGenkitWithModels 逻辑
```

**Rationale:**
- 将共享结构体放在 `providers.go`，因为它是 models 包的入口文件
- 各 Provider 文件保持独立，职责清晰
- `genkit.go` 聚合所有 Provider 的模型注册逻辑

### 5. Genkit ModelSupports 处理

当前代码中，每个 Provider 注册模型时都硬编码了 `ModelSupports`：

```go
&ai.ModelOptions{
    Supports: &ai.ModelSupports{
        Multiturn:  true,
        SystemRole: true,
        Tools:      true,
        Media:      true,  // 仅 Deepseek
    },
}
```

**决策:** 保持现状，暂时不根据配置动态设置。

**Rationale:**
- Genkit 的 `ModelSupports` 影响模型是否能正确处理特定请求类型
- 错误配置可能导致运行时错误
- 配置中的 `reasoning`、`vision` 字段预留用于未来应用层路由决策
- 如需动态配置，可在后续变更中谨慎实现

## Risks / Trade-offs

### Risk 1: Ollama 配置破坏性变更

**Risk:** 现有用户配置 `"models": ["llama2", "mistral"]` 将无法解析。

**Mitigation:**
- 在文档中明确标注 BREAKING CHANGE
- 提供迁移示例
- 考虑在 CHANGELOG 中高亮提示

### Risk 2: 配置错误导致无模型可用

**Risk:** 用户忘记配置 `models` 或配置格式错误，导致系统无任何模型可用。

**Mitigation:**
- 在 `NewGenkitWithModels` 中添加检测：如果所有 Provider 的 models 都为空，记录警告日志
- 文档中强调必须配置至少一个模型

### Risk 3: 元数据字段未被使用

**Risk:** 配置了大量元数据但代码中未使用，可能让用户困惑。

**Mitigation:**
- 文档中明确说明这些字段为"预留供日后使用"
- 配置示例中展示可选字段的使用方式

### Trade-off: 灵活性 vs 简洁性

**Trade-off:** 全部元数据字段可选让配置简洁，但可能导致用户不知道有哪些可配置项。

**Decision:** 选择简洁性，依赖文档和示例展示完整功能。

## Migration Plan

### 用户迁移步骤

1. **备份现有配置**:
   ```bash
   cp ~/.nfa/nfa.json ~/.nfa/nfa.json.backup
   ```

2. **更新 Ollama 配置**:
   ```json
   // 旧格式
   {
     "ollama": {
       "models": ["llama2", "mistral"]
     }
   }

   // 新格式
   {
     "ollama": {
       "models": [
         {"name": "llama2"},
         {"name": "mistral"}
       ]
     }
   }
   ```

3. **为 Deepseek 和 OpenAI Compatible 添加 models 配置**（如果之前没有配置过）

4. **验证配置**:
   ```bash
   nfa models list
   ```

### 代码迁移步骤

1. 定义 `ModelConfig` 和 `ModelCost` 结构体
2. 更新三个 `Options` 结构体
3. 修改 `RegisterModels` 方法实现
4. 更新 `NewGenkitWithModels` 逻辑
5. 更新文档

### 回滚策略

如果发现严重问题，可以：
1. 回滚代码变更
2. 用户恢复备份的配置文件
3. 重新发布 patch 版本

## Open Questions

**Q1: 是否需要在解析配置时验证 model 名称格式？**

A: 暂不验证。保持灵活性，允许用户自由命名。如果名称错误，Genkit 注册时会失败。

**Q2: 未来如何使用配置的元数据？**

A: 预留的元数据字段可用于：
- `reasoning`: 选择不同的生成策略（如 thinking model 的特殊处理）
- `vision`: 作为视觉任务的候选模型
- `cost`: 记录使用成本，生成成本报告
- `contextWindow`: 在对话历史过长时截断
- `maxOutputTokens`: 设置 max_tokens 参数限制输出

这些功能需要在独立的需求中设计和实现。
