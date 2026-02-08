# 数据提供商 (Data Providers)

NFA 支持通过配置数据提供商来获取实时金融数据和外部数据源，增强 Agent 的分析能力。

## 概述

数据提供商是 NFA 访问外部数据源的接口，主要包括：

- **Alpha Vantage** - 提供股票、外汇、加密货币等金融数据
- **腾讯云 WSA** - 提供网络搜索能力（详见 [网络搜索指南](web-search.md)）

通过配置数据提供商，Agent 可以在需要时主动获取最新的市场数据、财务报表、技术指标等信息。

## Alpha Vantage

Alpha Vantage 是一个免费的金融数据 API 服务，提供全球股票、外汇、加密货币等实时和历史数据。

### 配置方式

在 `~/.nfa/nfa.json` 的 `dataProviders` 数组中添加 Alpha Vantage 配置：

```json
{
  "modelProviders": [...],
  "defaultModels": {...},
  "dataProviders": [
    {
      "alphaVantage": {
        "apiKey": "your-alpha-vantage-api-key"
      }
    }
  ]
}
```

**配置参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `apiKey` | string | 是 | Alpha Vantage API 密钥 |

### 获取 API 密钥

1. 访问 [Alpha Vantage 官网](https://www.alphavantage.co/support/#api-key)
2. 免费注册账号
3. 获取免费 API 密钥

免费 API 密钥有调用频率限制（每分钟 5 次请求，每天 500 次请求）。如需更高配额，可购买高级版。

### 支持的数据类型

Alpha Vantage 通过 MCP (Model Context Protocol) 动态注册工具，支持的数据类型包括：

- **股票数据** - 实时报价、历史价格、技术指标
- **外汇数据** - 汇率查询、历史汇率
- **加密货币** - 实时价格、历史数据
- **基本面数据** - 公司财报、财务指标
- **经济指标** - GDP、CPI、利率等宏观数据

### 工具注册机制

Alpha Vantage 使用 MCP 协议，工具注册流程如下：

1. NFA 启动时创建 MCP 客户端
2. 客户端连接到 Alpha Vantage MCP 服务器
3. 自动获取可用工具列表
4. 注册为 NFA 的可用工具

这种方式的好处是：
- 自动发现新工具，无需手动更新
- 统一的工具调用接口
- 动态加载，灵活性高

### 使用场景

#### 场景 1：查询股票价格

```
用户：苹果公司（AAPL）现在的股价是多少？
```

Agent 会调用 Alpha Vantage 的实时报价工具，获取 AAPL 的最新股价。

#### 场景 2：获取历史数据

```
用户：帮我查询特斯拉过去一年的股价走势
```

Agent 会调用时间序列数据工具，获取 TSLA 的历史价格数据。

#### 场景 3：技术分析

```
用户：分析一下 Nvidia 的技术指标
```

Agent 会调用技术指标工具（如 RSI、MACD），获取 NVDA 的技术分析数据。

#### 场景 4：财务分析

```
用户：查询微软最近的财务报表数据
```

Agent 会调用财务数据工具，获取 MSFT 的营收、利润等基本面信息。

## 数据提供商工作流程

当 Agent 需要使用数据时，会按照以下流程操作：

```
用户提问
    ↓
Agent 判断是否需要外部数据
    ↓
选择合适的数据提供商和工具
    ↓
调用工具获取数据
    ↓
分析数据并回答用户问题
```

这个流程是自动的，用户无需手动指定使用哪个工具。

## API 调用注意事项

### 1. 调用频率限制

Alpha Vantage 免费版有调用频率限制：

- **每分钟**：5 次请求
- **每天**：500 次请求

如果超出限制，会收到错误提示。建议：
- 避免短时间内重复查询相同数据
- 合理利用缓存机制
- 考虑升级 API 计划

### 2. 数据延迟

不同数据类型的延迟可能不同：

- **实时报价**：可能有轻微延迟（15-20 分钟）
- **历史数据**：通常无延迟
- **基本面数据**：更新频率较低（季度/年度）

### 3. 数据准确性

- Alpha Vantage 数据来源可靠，但仍需交叉验证
- 市场数据可能有延迟，不应用于实时交易决策
- Agent 会结合多个数据源提高准确性

### 4. 错误处理

如果 API 调用失败，Agent 会：

1. 记录详细错误日志
2. 尝试使用其他数据源
3. 告知用户数据获取失败的原因

## API 密钥管理

### 安全建议

1. **不要泄露 API 密钥**
   - 不要将密钥提交到版本控制
   - 不要在公开场合分享密钥

2. **使用环境变量（可选）**
   - 可以考虑将密钥存储在环境变量中
   - 通过配置管理工具动态加载

3. **定期更换密钥**
   - 如果密钥可能已泄露，立即更换
   - 在 Alpha Vantage 控制台可以重新生成密钥

### 监控使用量

在 [Alpha Vantage 控制台](https://www.alphavantage.co/dashboard/) 可以查看：

- API 调用统计
- 剩余调用次数
- 使用峰值时段

根据使用情况选择合适的 API 计划。

## 与其他功能的配合

数据提供商可以与 NFA 的其他功能深度集成：

### 与网络搜索配合

```
用户：分析一下最近的科技股走势
```

Agent 可以：
1. 使用 Alpha Vantage 获取科技股的历史数据
2. 使用网络搜索获取相关的新闻和分析
3. 结合两者提供全面的分析

### 与网页浏览配合

```
用户：帮我查看苹果公司的财报详情
```

Agent 可以：
1. 使用 Alpha Vantage 获取关键财务数据
2. 使用 WebBrowse 访问财报原文
3. 深入分析具体指标

### 与自定义技能配合

你可以创建自定义技能，组合多个数据源：

```markdown
---
name: comprehensive-stock-analysis
description: 对股票进行全面分析
---

1. 使用 Alpha Vantage 获取股票的历史价格和技术指标
2. 使用网络搜索搜索该股票的最新新闻
3. 综合分析技术面、基本面和市场情绪
4. 给出投资建议（需说明这不构成投资建议）
```

## 配置示例

### 基础配置

只配置 Alpha Vantage：

```json
{
  "modelProviders": [
    {"ollama": {}}
  ],
  "defaultModels": {
    "main": "ollama/llama2",
    "fast": "ollama/llama2",
    "vision": ""
  },
  "dataProviders": [
    {
      "alphaVantage": {
        "apiKey": "YOUR_ALPHA_VANTAGE_API_KEY"
      }
    }
  ]
}
```

### 完整配置

同时配置 Alpha Vantage 和腾讯云 WSA：

```json
{
  "modelProviders": [
    {"ollama": {}},
    {
      "deepseek": {
        "apiKey": "your-deepseek-api-key"
      }
    }
  ],
  "defaultModels": {
    "main": "deepseek/deepseek-chat",
    "fast": "ollama/mistral",
    "vision": ""
  },
  "dataProviders": [
    {
      "alphaVantage": {
        "apiKey": "YOUR_ALPHA_VANTAGE_API_KEY"
      }
    },
    {
      "tcloudWSA": {
        "secretID": "YOUR_TENCENT_SECRET_ID",
        "secretKey": "YOUR_TENCENT_SECRET_KEY"
      }
    }
  ]
}
```

## 故障排查

### 数据提供商未生效

1. **检查配置文件**
   ```bash
   # 验证 JSON 格式
   cat ~/.nfa/nfa.json | jq .
   ```

2. **检查 API 密钥**
   - 确认密钥格式正确
   - 尝试在 Alpha Vantage 控制台测试密钥

3. **查看日志**
   ```bash
   tail -f ~/.nfa/nfa.log
   ```

### 遇到调用频率限制

1. **检查使用量**
   - 在 Alpha Vantage 控制台查看调用统计

2. **优化查询策略**
   - 避免重复查询相同数据
   - 使用历史数据而不是实时数据

3. **升级 API 计划**
   - 如需更高配额，可购买高级版 API 密钥

### 数据不完整或错误

1. **验证数据源**
   - 使用其他金融网站交叉验证数据

2. **检查股票代码**
   - 确认使用的股票代码正确（如 AAPL、TSLA）

3. **尝试不同的数据类型**
   - 某些股票可能不支持特定的数据类型

## 注意事项

1. **数据延迟** - 免费数据可能有延迟，不应用于实时交易决策

2. **免责声明** - NFA 提供的分析仅供参考，不构成投资建议

3. **数据准确性** - Agent 会尽力保证数据准确，但建议多源验证

4. **成本控制** - 合理使用 API，避免超出免费配额

5. **市场风险** - 金融投资有风险，决策需谨慎
