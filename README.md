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
- [自定义技能](docs/guides/skills.md) - 通过自定义技能扩展 NFA 能力
