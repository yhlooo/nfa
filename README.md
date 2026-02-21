# NFA (Not Financial Advice, 非财务建议)

基于大语言模型的金融交易顾问 Agent 。 **这不构成财务建议。**

Financial Trading LLM AI Agent. **This is Not Financial Advice.**

当前该项目还处于非常早期阶段

## 快速开始

在 `~/.nfa/nfa.json` 中添加配置。详见 [配置参考](docs/reference/config.md)。

通过以下命令可编译并运行：

```bash
go run ./cmd/nfa
```

## 文档

- [配置参考](docs/reference/config.md) - 配置文件结构和选项
- [命令行指南](docs/guides/command-line.md) - 命令行参数和交互式命令
- [自定义技能](docs/guides/skills.md) - 通过自定义技能扩展 NFA 能力

## 交互式命令

在交互式对话模式中，NFA 提供了强大的命令系统：

- `/model` - 动态切换模型（支持主模型、快速模型、视觉模型）
- `/clear` - 清空对话上下文
- `/summarize` - 生成对话摘要
- `/exit` - 退出程序

详见 [命令行指南](docs/guides/command-line.md#交互式命令)。
