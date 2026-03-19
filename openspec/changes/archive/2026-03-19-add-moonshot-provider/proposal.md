# 添加月之暗面 Moonshot AI 模型提供商支持

## 概述

为 NFA 添加对月之暗面（Moonshot AI）Kimi 模型的支持，使用户能够通过配置使用 kimi-k2.5 等模型。

## 背景

NFA 已支持多个模型提供商（Ollama、智谱 ZAI、通义千问 Qwen、DeepSeek、OpenAI 兼容）。月之暗面是国内领先的 AI 公司，其 Kimi 系列模型在中文理解和长文本处理方面表现出色。

## 目标

1. 添加 Moonshot AI 作为新的模型提供商
2. 支持 kimi-k2.5 模型
3. 遵循现有提供商的实现模式（参考 ZAI）
4. 支持 Moonshot API 的推理和视觉能力

## 范围

### 包含

- Moonshot 提供商的配置结构
- kimi-k2.5 模型的默认配置
- 模型注册和初始化逻辑
- 配置文档更新
- i18n 国际化支持

### 不包含

- 其他 Moonshot 模型的详细配置
- Moonshot 特有的高级 API 功能

## 技术方案

Moonshot API 兼容 OpenAI API 格式，可直接复用现有的 `oai.OpenAICompatible` 插件实现。

实现参考 ZAI 的模式：
- 定义 `MoonshotOptions` 结构体
- 提供默认的 Base URL 和模型列表
- 委托给 `OpenAICompatibleOptions.RegisterModels` 进行模型注册

## 成功标准

1. 用户可在配置文件中添加 Moonshot 提供商配置
2. 能够成功调用 kimi-k2.5 模型
3. 模型选择界面显示 Moonshot 模型
