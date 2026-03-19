## Why

当前 `dataProviders` 配置使用数组结构，但数据供应商每类只需要配置一个实例，数组形式既冗余又容易产生歧义。改为结构体形式可以使配置更简洁、语义更清晰。

## What Changes

- **BREAKING**: 将 `dataProviders` 从数组 `[]DataProvider` 改为结构体 `DataProviders`
- 重命名类型 `DataProvider` 为 `DataProviders`（复数形式）
- 更新配置文件结构，JSON 配置格式从数组变为对象

配置格式变化示例：
```json
// 旧格式
{
  "dataProviders": [
    { "alphaVantage": { "apiKey": "..." } },
    { "tcloudWSA": { "secretID": "...", "secretKey": "..." } }
  ]
}

// 新格式
{
  "dataProviders": {
    "alphaVantage": { "apiKey": "..." },
    "tcloudWSA": { "secretID": "...", "secretKey": "..." }
  }
}
```

## Capabilities

### New Capabilities

无新增能力。

### Modified Capabilities

- `config`: 配置文件结构变更，`dataProviders` 字段格式从数组改为结构体

## Impact

- **代码文件**:
  - `pkg/agents/agent.go`: 重命名类型，修改字段类型
  - `pkg/configs/config.go`: 修改 `DataProviders` 字段类型
  - `pkg/agents/genkit.go`: 修改遍历逻辑
- **文档**:
  - `docs/reference/config.md`: 更新配置示例
- **用户影响**: 用户需要更新现有的 `~/.nfa/nfa.json` 配置文件格式
