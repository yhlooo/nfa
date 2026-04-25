# 模型效果评分展示 - 实现任务

## 1. ModelConfig 新增 Score 字段

- [x] 1.1 在 `pkg/models/providers.go` 的 `ModelConfig` 结构体中添加 `Score int` 字段，JSON tag `"score,omitempty"`

## 2. i18n 表头

- [x] 2.1 在 `pkg/commands/i18n.go` 中添加 `MsgScoreTag` i18n 消息，ID `"commands.ScoreTag"`，英文 `"Score"`

## 3. 评分展示逻辑

- [x] 3.1 在 `pkg/commands/models.go` 中实现 `scoreToStars(score int) string` 函数
- [x] 3.2 修改 `outputModelList` 函数：表头加 Score 列、对齐数组加 `Center`、每行数据加星标字符串

## 4. 代码质量

- [x] 4.1 运行 `go fmt ./...` 格式化代码
- [x] 4.2 运行 `go vet ./...` 检查代码问题
- [x] 4.3 运行 `go test ./...` 确保测试通过

## 5. 手动验证

- [x] 5.1 构建运行 `go run ./cmd/nfa models list` 确认 Score 列正确展示
- [x] 5.2 验证 Score=0 的模型不显示评分
