## 1. 新增 history 包

- [x] 1.1 创建 `pkg/history/history.go`，定义 `Entry` 和 `History` 结构体
- [x] 1.2 在 `History` 结构体中实现 `Add(content string)` 方法
- [x] 1.3 在 `History` 结构体中实现 `Up() string` 方法（向上浏览）
- [x] 1.4 在 `History` 结构体中实现 `Down() string` 方法（向下浏览）
- [x] 1.5 在 `History` 结构体中实现 `ResetNav()` 方法（重置导航状态）
- [x] 1.6 创建 `pkg/history/load.go`，实现 `LoadHistory(path string) (*History, error)` 函数
- [x] 1.7 在 `pkg/history/load.go` 中实现 `SaveHistory(path string, h *History) error` 函数

## 2. UI 层 - InputBox 扩展

- [x] 2.1 在 `pkg/ui/chat/input.go` 的 `InputBox` 结构体中添加字段：
  - `history *history.History`
  - `historyPath string`
  - `historyIndex int`
  - `historyTempValue string`
- [x] 2.2 修改 `NewInputBox` 函数签名，接受 history 和 historyPath 参数
- [x] 2.3 在 `Update` 方法的 `tea.KeyUp` 分支中添加历史浏览逻辑（仅单行模式）
- [x] 2.4 在 `Update` 方法的 `tea.KeyDown` 分支中添加历史浏览逻辑（仅单行模式）
- [x] 2.5 在 `Reset` 方法中重置 `historyIndex` 和 `historyTempValue`

## 3. UI 层 - ChatUI 集成

- [x] 3.1 在 `pkg/ui/chat/ui.go` 的 `Run` 方法中，初始化时加载历史记录
- [x] 3.2 修改 `NewInputBox` 调用，传入 history 和 historyPath
- [x] 3.3 在 `updateInInputState` 方法中，提交输入后调用 `history.Add` 和 `SaveHistory`
- [x] 3.4 添加 history 相关 import

## 4. 历史文件路径处理

- [x] 4.1 确定历史文件路径（与配置文件同目录：`~/.nfa/history.json`）
- [x] 4.2 确保目录存在，必要时创建

## 5. 测试和验证

- [x] 5.1 手动测试：输入多条内容，验证 `history.json` 正确生成
- [x] 5.2 手动测试：↑ 键能正确浏览历史
- [x] 5.3 手动测试：↓ 键能正确浏览历史
- [x] 5.4 手动测试：浏览到底后按 ↓ 恢复当前输入
- [x] 5.5 手动测试：多行模式下 ↑/↓ 不触发历史浏览
- [x] 5.6 手动测试：空输入不被记录
- [x] 5.7 手动测试：重启程序后历史记录保留
- [x] 5.8 手动测试：超过 100 条时旧记录被删除

## 6. 代码质量

- [x] 6.1 运行 `go fmt ./...`
- [x] 6.2 运行 `go vet ./...`
- [x] 6.3 运行 `go test ./...`
