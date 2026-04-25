## Context

当前 `pkg/commands/models.go` 只实现了 `newModelsListCommand`，`newModelsCommand` 中仅注册了该子命令。

配置系统 `pkg/configs` 已有 `LoadConfig` 和 `SaveConfig` 两个函数，支持从 JSON 文件读写完整配置。配置在 `root.go` 的 `PersistentPreRunE` 中加载并注入到 context，通过 `configs.ConfigFromContext(ctx)` 和 `configs.ConfigPathFromContext(ctx)` 访问。

模型供应商通过 `models.ModelProvider` 结构体表示，该结构体是一个"oneof"类型——各供应商字段通过 `omitempty` 区分，同一时间只有一个字段有值。供应商选项结构体的字段各有不同，主要集中在 API 密钥、BaseURL 等方面。

## Goals / Non-Goals

**Goals:**
- 通过 `nfa models add <provider>` 命令将供应商配置追加/覆盖写入 `nfa.json`
- 覆盖所有 8 种已有供应商类型
- 使用平铺的 CLI flag（所有 flag 统一注册，按需校验）
- 重复添加时覆盖已有配置（而非报错）

**Non-Goals:**
- 不做连接验证
- 不增加删除、修改其他配置的子命令
- 不修改 `ModelProvider` 等现有数据结构
- 不为供应商类型动态注册 flag（所有 flag 平铺在每个子命令中）

## Decisions

### 决策 1：使用单一 `add` 子命令 + 位置参数指定供应商名

```
nfa models add <provider> [flags]
```

而非为每个供应商创建独立子命令（如 `nfa models add-deepseek`）。

**理由**：命令数量不会随供应商增加而膨胀；用户心智模型是"添加一个供应商"而不是"调用 deepseek 专属添加命令"。

### 决策 2：使用已知供应商映射表

在 `pkg/commands/models.go` 中维护一个 `supportedProviders` map，将用户输入的供应商名映射到对应的 JSON 字段名和选项构建逻辑：

```go
var supportedProviders = map[string]providerInfo{
    "deepseek":          {jsonField: "deepseek", ...},
    "qwen":              {jsonField: "qwen", ...},
    "moonshot":          {jsonField: "moonshotai", ...},
    ...
}
```

**理由**：与 `ModelProvider` 的 JSON tag 解耦，用户输入可用更友好的名称（如 `moonshot` 而非 `moonshotai`）。

### 决策 3：覆盖策略

若同类型供应商已存在配置，直接用新配置覆盖。

**实现方式**：遍历 `cfg.ModelProviders`，若找到同类型则替换；若未找到则追加。

**理由**：用户意图明确，覆盖比报错更符合"配置一个可用的供应商"的直觉。若有更复杂需求（如多实例），当前 `ModelProvider` 结构不支持，超出本需求范围。

### 决策 4：所有 flag 平铺注册

不管供应商类型，所有 flag（`--apiKey`、`--baseURL`、`--name`、`--serverAddress`、`--timeout`）一次性注册在 `add` 子命令上。`RunE` 中根据供应商类型校验必填字段是否填写。

**理由**：cobra 的 flag 注册是静态的，平铺避免了在 `PreRunE` 中动态注册的复杂性。未使用的 flag 用户看不见/不填即可。

### 决策 5：saveConfig 复用现有关口

`configs.SaveConfig` 已实现读-改-写模式。`add` 子命令将使用 `ConfigPathFromContext(ctx)` 获取路径，在内存中修改配置后调用 `SaveConfig` 写入。

**理由**：复用现有基础设施，避免重新实现序列化逻辑。

## Risks / Trade-offs

- **flag 膨胀风险**：未来若新增供应商类型且参数不同，flag 列表会增长。但目前 5 个 flag 在可接受范围内。
- **覆盖静默**：覆盖已有配置时无额外提示，用户可能无感知。可考虑输出确认信息。
- **openai-compatible 的 `--name` 歧义**：`--name` 同时是供应商展示名和 cobra 的内置 flag，但 cobra 不注册 `name` flag，不冲突。
