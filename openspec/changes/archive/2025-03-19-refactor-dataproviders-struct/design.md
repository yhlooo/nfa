## Context

当前配置系统中，`dataProviders` 使用数组结构：

```go
type DataProvider struct {
    AlphaVantage    *alphavantage.Options             `json:"alphaVantage,omitempty"`
    TencentCloudWSA *websearch.TencentCloudWSAOptions `json:"tcloudWSA,omitempty"`
}

DataProviders []DataProvider  // 数组
```

这种设计与实际使用场景不符：
- 数据供应商每类只需要配置一个实例
- 数组形式导致配置冗余（每项只有一个非空字段）
- `modelProviders` 需要数组是因为可以配置多个同类供应商（如多个 Ollama 实例），但 `dataProviders` 不需要

## Goals / Non-Goals

**Goals:**
- 将 `dataProviders` 从数组改为结构体，使配置更简洁直观
- 保持与 `modelProviders` 的设计差异（模型可多实例，数据源单实例）

**Non-Goals:**
- 不修改 `modelProviders` 的数组结构
- 不添加新的数据供应商类型

## Decisions

### 决策 1：类型重命名

将 `DataProvider`（单数）重命名为 `DataProviders`（复数），表示这是一个包含多种数据供应商的结构体。

**理由**：原名 `DataProvider` 会被误解为单个供应商配置，新名称更清晰地表达"数据供应商集合"的含义。

### 决策 2：字段类型变更

```go
// 旧结构
type DataProvider struct {
    AlphaVantage    *alphavantage.Options             `json:"alphaVantage,omitempty"`
    TencentCloudWSA *websearch.TencentCloudWSAOptions `json:"tcloudWSA,omitempty"`
}

// 新结构
type DataProviders struct {
    AlphaVantage    *alphavantage.Options             `json:"alphaVantage,omitempty"`
    TencentCloudWSA *websearch.TencentCloudWSAOptions `json:"tcloudWSA,omitempty"`
}
```

字段定义保持不变，只是外层容器从 `[]DataProvider` 变为 `DataProviders`。

### 决策 3：遍历逻辑调整

```go
// 旧逻辑
for _, p := range a.dataProviders {
    switch {
    case p.AlphaVantage != nil:
        // ...
    case p.TencentCloudWSA != nil:
        // ...
    }
}

// 新逻辑
if a.dataProviders.AlphaVantage != nil {
    // ...
}
if a.dataProviders.TencentCloudWSA != nil {
    // ...
}
```

新逻辑更简洁，直接检查各字段是否配置。

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 用户需要手动更新配置文件格式 | 在文档中提供清晰的迁移指南 |
| 旧配置文件解析失败 | Go 的 JSON 解析会忽略多余字段，不会报错，但新字段为空 |

## Migration Plan

1. 发布新版本，包含此变更
2. 用户更新后需要修改 `~/.nfa/nfa.json` 中的 `dataProviders` 格式
3. 更新文档 `docs/reference/config.md` 中的所有示例
