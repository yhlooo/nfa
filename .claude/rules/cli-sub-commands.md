---
paths:
  - "pkg/commands/*.go"
---

# CLI 子命令规范

## 子命令与文件结构映射

- 根命令定义在 `pkg/commands/root.go` 中
- 每个一级子命令（比如 `nfa version ...` `nfa models ...`）及其子命令应至少在 `pkg/commands/` 目录下有一个文件，以一级子命令名命名
- 如果一级子命令下的子命令较多（多于 3 个），也可为二级子命令添加独立文件，比如可以在 `models_list.go` 实现 `nfa models list` 子命令

## 子命令定义

每个子命令定义都类似以下内容（可以参考 `pkg/commands/otter.go` ）：

- 以子命令命名的 `XxxOptions` 结构体，定义子命令绑定到命令行 flag 的选项。需要有一个 `AddPFlags(fs *pflag.FlagSet)` 成员方法，用于把结构体字段绑定到命令行 flag
- 方法 `NewXxxOptions() XxxOptions` 用于创建带默认值的 `XxxOptions`
- 以子命令命名的 `newXxxCommand() *cobra.Command` 用于创建子命令的 `*cobra.Command` 对象， `*cobra.Command` 对象应包含以下字段：
  - `Use` 子命令用法示例，子命令名和位置参数占位符
  - `Short` 子命令的简短描述
  - `Args` 位置参数数量定义
  - `RunE` （可选）执行该命令的逻辑
