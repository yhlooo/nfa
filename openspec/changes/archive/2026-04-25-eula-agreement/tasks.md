## 1. 创建 pkg/eula 模块

- [x] 1.1 创建 `pkg/eula/eula.md`，撰写初始协议内容（后由用户调整）
- [x] 1.2 创建 `pkg/eula/eula.go`，实现核心逻辑：embed 协议文件、首次写入、SHA256 计算、签名文件读写、交互提示
- [x] 1.3 创建 `pkg/eula/i18n.go`，定义 EULA 相关提示的 `i18n.Message`（同意/拒绝询问、协议更新提示、输入无效提示等）
- [x] 1.4 创建 `pkg/eula/eula_test.go`，编写 EULA 逻辑的单元测试

## 2. 集成到启动流程

- [x] 2.1 在 `pkg/commands/root.go` 的 `PersistentPreRunE` 中，i18n 初始化之后，插入 EULA 检查调用（仅对根命令生效）
- [x] 2.2 确保子命令（`version`、`models list`、`otter` 等）不触发 EULA 检查

## 3. 国际化与翻译

- [x] 3.1 运行 `i18n-translate` skill 从代码中提取消息并更新翻译文件

## 4. 代码质量检查

- [x] 4.1 运行 `go fmt ./...` 格式化代码
- [x] 4.2 运行 `go vet ./...` 静态分析
- [x] 4.3 运行 `go test ./...` 确认所有测试通过（`pkg/agents` 的 `TestNewAgentSystemPrompt` 是已存在的无关失败）
