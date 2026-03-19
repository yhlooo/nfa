## 1. 类型定义修改

- [x] 1.1 在 `pkg/agents/agent.go` 中将 `DataProvider` 重命名为 `DataProviders`
- [x] 1.2 修改 `Options.DataProviders` 字段类型从 `[]DataProvider` 改为 `DataProviders`
- [x] 1.3 修改 `NFAAgent.dataProviders` 字段类型从 `[]DataProvider` 改为 `DataProviders`
- [x] 1.4 修改 `pkg/configs/config.go` 中 `Config.DataProviders` 字段类型

## 2. 逻辑调整

- [x] 2.1 修改 `pkg/agents/genkit.go` 中的数据供应商遍历逻辑，从 for-range 改为直接字段判断

## 3. 文档更新

- [x] 3.1 更新 `docs/reference/config.md` 中所有 `dataProviders` 配置示例

## 4. 验证

- [x] 4.1 运行 `go fmt ./...` 格式化代码
- [x] 4.2 运行 `go vet ./...` 静态分析
- [x] 4.3 运行 `go test ./...` 单元测试
