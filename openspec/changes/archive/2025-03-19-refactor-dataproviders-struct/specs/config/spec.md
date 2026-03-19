## ADDED Requirements

### Requirement: 数据供应商配置结构

系统 SHALL 使用结构体形式配置数据供应商，每种数据供应商类型最多配置一个实例。

配置格式：
```json
{
  "dataProviders": {
    "alphaVantage": {
      "apiKey": "your-api-key"
    },
    "tcloudWSA": {
      "secretID": "your-secret-id",
      "secretKey": "your-secret-key"
    }
  }
}
```

#### Scenario: 配置单个数据供应商

- **WHEN** 用户在配置文件中只配置 `alphaVantage`
- **THEN** 系统仅加载 Alpha Vantage 数据工具，不加载其他数据工具

#### Scenario: 配置多个数据供应商

- **WHEN** 用户在配置文件中同时配置 `alphaVantage` 和 `tcloudWSA`
- **THEN** 系统同时加载 Alpha Vantage 和腾讯云搜索工具

#### Scenario: 不配置任何数据供应商

- **WHEN** 用户未在配置文件中配置 `dataProviders` 或配置为空对象
- **THEN** 系统正常启动，不加载任何数据供应商工具

### Requirement: Alpha Vantage 配置

系统 SHALL 支持通过 `alphaVantage` 字段配置 Alpha Vantage 数据供应商。

配置字段：
- `apiKey`（必填）：API 密钥

#### Scenario: 配置 Alpha Vantage

- **WHEN** 用户配置 `dataProviders.alphaVantage.apiKey` 为有效值
- **THEN** 系统注册 Alpha Vantage 相关工具，可查询股票行情等数据

### Requirement: 腾讯云搜索配置

系统 SHALL 支持通过 `tcloudWSA` 字段配置腾讯云 Web Search Agent 数据供应商。

配置字段：
- `secretID`（必填）：腾讯云 Secret ID
- `secretKey`（必填）：腾讯云 Secret Key
- `endpoint`（可选）：服务端点

#### Scenario: 配置腾讯云搜索

- **WHEN** 用户配置 `dataProviders.tcloudWSA.secretID` 和 `secretKey` 为有效值
- **THEN** 系统注册腾讯云搜索工具，可执行联网搜索
