# Web Browse Tool Specification

## Requirements

### Requirement: WebBrowse 工具超时参数

WebBrowse 工具 SHALL 支持通过 `timeout` 参数设置执行超时时间。

参数定义：
- 字段名：`timeout`
- 类型：整数
- 单位：秒
- 默认值：60
- 可选：是

#### Scenario: 使用默认超时

- **GIVEN** Agent 调用 WebBrowse 工具时未提供 `timeout` 参数
- **WHEN** 工具执行
- **THEN** 工具在 60 秒后超时

#### Scenario: 使用自定义超时

- **GIVEN** Agent 调用 WebBrowse 工具时提供 `timeout` 参数为 30
- **WHEN** 工具执行
- **THEN** 工具在 30 秒后超时

#### Scenario: 超时返回错误

- **GIVEN** WebBrowse 工具执行超过设定超时时间
- **WHEN** 超时发生
- **THEN** 工具返回错误，错误信息包含超时相关描述

#### Scenario: 超时覆盖所有操作

- **GIVEN** WebBrowse 工具正在执行任一阶段（页面加载、文本提取、截图、视觉理解）
- **WHEN** 达到超时时间
- **THEN** 当前操作被中断，工具返回超时错误
