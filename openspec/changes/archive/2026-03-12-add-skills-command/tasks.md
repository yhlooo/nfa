## 1. Agent Layer - Meta 扩展

- [x] 1.1 在 `pkg/agents/meta.go` 中添加常量 `MetaKeySkills = "skills"`
- [x] 1.2 在 `pkg/agents/meta.go` 中添加 `GetMetaSkillsValue(meta map[string]any) []skills.SkillMeta` 函数
- [x] 1.3 修改 `pkg/agents/acp.go` 的 `Initialize` 方法，在返回的 Meta 中添加技能列表

## 2. UI Layer - 数据结构

- [x] 2.1 在 `pkg/ui/chat/ui.go` 的 `ChatUI` 结构体中添加 `skills []skills.SkillMeta` 字段

## 3. UI Layer - 数据接收

- [x] 3.1 修改 `pkg/ui/chat/acp.go` 的 `initAgent` 方法，接收技能列表并存储到 `ui.skills`

## 4. UI Layer - 命令注册

- [x] 4.1 修改 `pkg/ui/chat/ui.go` 的 `Run` 方法，在 `NewInputBox` 的命令选项中添加 `{Name: "skills", Description: "List loaded skills"}`

## 5. UI Layer - 命令处理

- [x] 5.1 在 `pkg/ui/chat/ui.go` 的 `updateInInputState` 方法中添加 `/skills` 命令检测
- [x] 5.2 实现 `printSkillsList() tea.Cmd` 方法，格式化输出技能列表

## 6. 测试和验证

- [x] 6.1 手动测试：输入 `/skills` 显示技能列表
- [x] 6.2 手动测试：验证 builtin 和 local 技能分组正确
- [x] 6.3 手动测试：验证技能数量统计正确
- [x] 6.4 手动测试：只有 builtin 技能时不显示 Local skills 分组
- [x] 6.5 手动测试：只有 local 技能时不显示 Builtin skills 分组
- [x] 6.6 手动测试：没有技能时显示 "0 skills"

## 7. 代码质量

- [x] 7.1 运行 `go fmt ./...`
- [x] 7.2 运行 `go vet ./...`
- [x] 7.3 运行 `go test ./...`
