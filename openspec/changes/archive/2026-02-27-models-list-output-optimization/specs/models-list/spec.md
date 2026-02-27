# models-list Specification

## Purpose
定义 `nfa models list` 命令的行为，用于列出和展示所有可用模型的信息。

## Requirements

### Requirement: List all available models
系统 MUST 支持列出所有配置的模型提供商中的可用模型。

#### Scenario: Execute models list command
- **WHEN** 用户执行 `nfa models list`
- **THEN** 系统显示所有可用模型
- **AND** 每个模型独占一行
- **AND** 输出按提供商名称分组和排序

### Requirement: Human-readable text output
系统 MUST 以人类可读的文本格式输出模型信息。

#### Scenario: Display model with capabilities
- **WHEN** 模型具有推理能力（Reasoning == true）
- **THEN** 输出括号包含 "Reasoning" 文字
- **AND** 括号及内容使用弱化/淡色显示

#### Scenario: Display model with vision capability
- **WHEN** 模型具有视觉能力（Vision == true）
- **THEN** 输出括号包含 "Visual" 文字
- **AND** 括号及内容使用弱化/淡色显示

#### Scenario: Display model with both capabilities
- **WHEN** 模型同时具有推理和视觉能力
- **THEN** 输出括号包含 "Reasoning, Visual" 文字
- **AND** 两个能力用 ", " 分隔
- **AND** 括号及内容使用弱化/淡色显示

#### Scenario: Display model without capabilities
- **WHEN** 模型没有任何特殊能力
- **THEN** 括号内只包含上下文大小（如 "(128K)"）
- **AND** 括号及内容使用弱化/淡色显示

#### Scenario: Display context window size in parentheses
- **WHEN** 显示模型信息
- **THEN** 上下文窗口大小 MUST 显示在括号内
- **AND** 格式化为人类可读形式（如 128000 → 128K）
- **AND** 位置在能力列表之后，用 ", " 分隔
- **AND** 示例：`(Reasoning, 128K)` 或 `(128K)`

#### Scenario: Display model description
- **WHEN** 模型有描述信息（Description 不为空）
- **THEN** 描述 MUST 显示在行尾
- **AND** 描述超过 50 字符时截断并添加 "..."
- **AND** 使用终端默认颜色

#### Scenario: No description available
- **WHEN** 模型没有描述信息
- **THEN** 描述位置显示 "-"

### Requirement: Output layout and alignment
系统 MUST 使用一致的列布局输出模型信息。

#### Scenario: Standard layout
- **WHEN** 输出模型列表
- **THEN** 使用以下列布局：
  - 列 1: 模型名称（左对齐，35 字符）
  - 列 2: 括号内容（左对齐，30 字符）
  - 列 3: 描述（左对齐，最大宽度 50 字符）

#### Scenario: Parentheses content format
- **WHEN** 构建括号内容
- **THEN** 格式为 `(能力列表, 上下文大小)`
- **AND** 能力和上下文用 ", " 分隔
- **AND** 多个能力用 ", " 分隔
- **AND** 示例：
  - `(Reasoning, 128K)` - 仅推理能力
  - `(Reasoning, Visual, 128K)` - 推理和视觉能力
  - `(128K)` - 无能力

#### Scenario: Long model name handling
- **WHEN** 模型名称超过 35 字符
- **THEN** 括号内容换行到下一行
- **AND** 描述保持在括号内容之后

#### Scenario: Format context window size
- **WHEN** 上下文窗口大小 >= 1024
- **THEN** 格式为 "{N}K"（如 128000 → 128K）
- **WHEN** 上下文窗口大小 < 1024
- **THEN** 直接显示数字（如 512 → 512）

### Requirement: Provider filtering
系统 MUST 支持按提供商过滤模型列表。

#### Scenario: Filter by provider name
- **WHEN** 用户执行 `nfa models list --provider deepseek`
- **THEN** 仅显示提供商为 "deepseek" 的模型
- **AND** 提供商名称匹配不区分大小写

#### Scenario: No models match provider filter
- **WHEN** 提供商过滤没有匹配结果
- **THEN** 显示 "No models found for provider: {name}"
- **AND** 退出码为 0

#### Scenario: Invalid provider name
- **WHEN** 用户执行 `nfa models list --provider nonexistent`
- **THEN** 显示 "No models found for provider: nonexistent"
- **AND** 不报错，正常退出

### Requirement: Capability filtering
系统 MUST 支持按能力过滤模型列表。

#### Scenario: Filter by reasoning capability
- **WHEN** 用户执行 `nfa models list --capability reasoning`
- **THEN** 仅显示具有推理能力的模型（Reasoning == true）

#### Scenario: Filter by vision capability
- **WHEN** 用户执行 `nfa models list --capability vision`
- **THEN** 仅显示具有视觉能力的模型（Vision == true）

#### Scenario: No models match capability filter
- **WHEN** 能力过滤没有匹配结果
- **THEN** 显示 "No models found with capability: {name}"
- **AND** 退出码为 0

#### Scenario: Invalid capability name
- **WHEN** 用户执行 `nfa models list --capability invalid`
- **THEN** 显示错误 "Invalid capability: invalid. Valid options: reasoning, vision"
- **AND** 退出码非 0

### Requirement: Combined filtering
系统 MUST 支持组合使用多个过滤条件。

#### Scenario: Filter by both provider and capability
- **WHEN** 用户执行 `nfa models list --provider zhipu --capability vision`
- **THEN** 仅显示智谱提供商的视觉模型
- **AND** 两个过滤条件使用 AND 逻辑

### Requirement: JSON output format
系统 MUST 支持以 JSON 格式输出模型信息，便于脚本解析。

#### Scenario: Output JSON format
- **WHEN** 用户执行 `nfa models list --format json`
- **THEN** 输出 JSON 对象
- **AND** JSON 结构如下：
  ```json
  {
    "models": [
      {
        "name": "provider/model-name",
        "provider": "provider",
        "description": "模型描述",
        "capabilities": {
          "reasoning": true,
          "vision": false
        },
        "contextWindow": 128000,
        "maxOutputTokens": 64000,
        "cost": {
          "input": 0.002,
          "output": 0.003,
          "cached": 0.0
        }
      }
    ]
  }
  ```

#### Scenario: JSON with filters
- **WHEN** 用户执行 `nfa models list --format json --provider deepseek`
- **THEN** JSON 输出仅包含过滤后的模型
- **AND** 结构与无过滤时相同

### Requirement: Terminal capability detection
系统 MUST 自动检测终端能力并调整输出。

#### Scenario: Disable colors with NO_COLOR
- **WHEN** 设置了 `NO_COLOR` 环境变量
- **THEN** 输出使用纯文本，不包含颜色
- **AND** emoji 仍然显示（如果终端支持）

#### Scenario: Non-terminal output
- **WHEN** 输出重定向到文件（如 `nfa models list > models.txt`）
- **THEN** 输出使用纯文本
- **AND** 不包含 ANSI 颜色转义序列

#### Scenario: Detect terminal color support
- **WHEN** 终端支持 TrueColor
- **THEN** 使用完整的颜色范围
- **WHEN** 终端仅支持 ANSI 256 色
- **THEN** 使用 256 色模式
- **WHEN** 终端仅支持基本 ANSI
- **THEN** 使用基本颜色

### Requirement: Provider name extraction
系统 MUST 从模型名称中提取提供商名称用于显示和过滤。

#### Scenario: Extract provider from model name
- **WHEN** 模型名称为 "deepseek/deepseek-reasoner"
- **THEN** 提供商名称为 "deepseek"
- **WHEN** 模型名称为 "aliyun/qwen3-max"
- **THEN** 提供商名称为 "aliyun"

### Requirement: Command help and usage
系统 MUST 提供清晰的命令帮助信息。

#### Scenario: Display help
- **WHEN** 用户执行 `nfa models list --help`
- **THEN** 显示命令描述
- **AND** 显示所有可用选项
- **AND** 显示使用示例

#### Scenario: Display usage examples in help
- **WHEN** 用户查看帮助信息
- **THEN** 包含以下示例：
  - `nfa models list` - 列出所有模型
  - `nfa models list --provider deepseek` - 仅显示 DeepSeek 模型
  - `nfa models list --capability reasoning` - 仅显示推理模型
  - `nfa models list --format json` - JSON 格式输出