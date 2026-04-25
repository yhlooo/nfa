## 1. 定义 i18n 消息与供应商映射

- [x] 1.1 在 `pkg/commands/i18n.go` 中新增 i18n 消息定义（子命令描述、错误提示、成功提示等）
- [x] 1.2 在 `pkg/commands/models.go` 中定义 `supportedProviders` 映射表，关联用户输入的供应商名、JSON 字段名、必填参数列表

## 2. 实现 `models add` 子命令

- [x] 2.1 实现 `newModelsAddCommand()` 函数，创建 cobra 子命令并注册全部平铺 flag（`--apiKey`、`--baseURL`、`--name`、`--serverAddress`、`--timeout`）
- [x] 2.2 实现 `runModelsAdd()` 函数：解析供应商名、校验必填参数、构建 `ModelProvider`、覆盖或追加、保存配置
- [x] 2.3 在 `newModelsCommand()` 中注册 `add` 子命令

## 3. 验证

- [x] 3.1 运行 `go fmt ./...` 格式化代码
- [x] 3.2 运行 `go vet ./...` 静态分析
- [x] 3.3 运行 `go test ./...` 单元测试
