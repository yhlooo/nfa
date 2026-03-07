# 重命名模型标识符 - 实现任务

## 1. Go 代码变更

### 1.1 核心数据结构
- [x] 1.1.1 修改 `pkg/models/model_routing.go` 中的 `Models` 结构体：`Main` → `Primary`, `Fast` → `Light`
- [x] 1.1.2 修改 `pkg/models/model_routing.go` 中的方法：`GetMain()` → `GetPrimary()`, `GetFast()` → `GetLight()`
- [x] 1.1.3 更新 `Models` 结构体的 JSON 标签：`` `json:"main"` `` → `` `json:"primary"` ``, `` `json:"fast"` `` → `` `json:"light"` ``

### 1.2 命令行参数
- [x] 1.2.1 修改 `pkg/commands/root.go`：将 `--fast-model` 参数改为 `--light-model`
- [x] 1.2.2 修改 `pkg/commands/root.go`：将 `DefaultFastModel` 字段改为 `DefaultLightModel`
- [x] 1.2.3 验证 `--model` 参数保持不变，映射到 Primary 模型

### 1.3 UI 常量和交互命令
- [x] 1.3.1 修改 `pkg/ui/chat/ui.go` 中的常量：`ModelTypeMain` → `ModelTypePrimary`, `ModelTypeFast` → `ModelTypeLight`
- [x] 1.3.2 修改 `pkg/ui/chat/ui.go` 中的命令处理：`/model :main` → `/model :primary`, `/model :fast` → `/model :light`
- [x] 1.3.3 更新 `pkg/ui/chat/ui.go` 中的菜单标题：`"Select main model"` → `"Select primary model"`, `"Select fast model"` → `"Select light model"`
- [x] 1.3.4 更新 `pkg/ui/chat/ui.go` 中的成功消息：`"Main model set to"` → `"Primary model set to"`, `"Fast model set to"` → `"Light model set to"`
- [x] 1.3.5 更新 `pkg/ui/chat/ui.go` 中的模型类型显示：`"(main)"` → `"(primary)"`, `"(fast)"` → `"(light)"`

### 1.4 Agent 元数据
- [x] 1.4.1 修改 `pkg/agents/meta.go` 中的元数据键：`v["main"]` → `v["primary"]`, `v["fast"]` → `v["light"]`
- [x] 1.4.2 检查并更新所有使用模型元数据的代码，确保引用新的键名

### 1.5 其他可能的引用
- [x] 1.5.1 使用全局搜索查找代码中所有 `"main"` 和 `"fast"` 字符串引用
- [x] 1.5.2 检查并更新测试文件中的相关引用
- [x] 1.5.3 检查并更新示例代码中的相关引用

### 1.6 代码质量检查
- [x] 1.6.1 运行 `go fmt ./...` 格式化所有代码
- [x] 1.6.2 运行 `go vet ./...` 检查代码问题
- [x] 1.6.3 运行 `go test ./...` 确保所有测试通过

## 2. 文档更新

### 2.1 用户文档
- [x] 2.1.1 更新 `docs/reference/config.md`：将 `main`/`fast` 改为 `primary`/`light`
- [x] 2.1.2 更新 `docs/reference/config.md` 中的所有配置示例
- [x] 2.1.3 更新 `docs/guides/command-line.md` 中的 CLI 参数说明：`--fast-model` → `--light-model`
- [x] 2.1.4 更新 `docs/guides/command-line.md` 中的所有命令示例
- [x] 2.1.5 更新 `docs/guides/model-config.md` 中的模型配置说明

### 2.2 项目文档
- [x] 2.2.1 更新 `CLAUDE.md` 中的模型配置说明
- [x] 2.2.2 检查并更新其他项目文档中的相关引用

### 2.3 规格文档
- [x] 2.3.1 更新 `openspec/specs/model-selection/spec.md`：将所有 `main`/`fast` 引用改为 `primary`/`light`
- [x] 2.3.2 更新 `openspec/specs/model-config/spec.md` 中的配置字段说明
- [x] 2.3.3 验证所有 scenario 的 WHEN/THEN 步骤使用新的命名

## 3. 验证和测试

### 3.1 功能测试
- [ ] 3.1.1 测试交互式命令：`/model` 和 `/model :primary` 能正确打开主模型选择菜单
- [ ] 3.1.2 测试交互式命令：`/model :light` 能正确打开轻量模型选择菜单
- [ ] 3.1.3 测试直接设置模型：`/model ollama/llama3.2` 能正确设置 primary 模型
- [ ] 3.1.4 测试直接设置模型：`/model :light ollama/mistral` 能正确设置 light 模型
- [ ] 3.1.5 测试 CLI 参数：`--model` 和 `--light-model` 能正确传递模型配置
- [ ] 3.1.6 测试配置文件加载：能正确读取 `defaultModels.primary` 和 `defaultModels.light`
- [ ] 3.1.7 测试配置文件保存：模型选择后能正确保存到新的字段名

### 3.2 UI 显示验证
- [ ] 3.2.1 验证模型选择菜单标题显示正确（"Select primary model", "Select light model"）
- [ ] 3.2.2 验证成功提示消息显示正确（"Primary model set to", "Light model set to"）
- [ ] 3.2.3 验证模型类型显示正确（"(primary)", "(light)"）
- [ ] 3.2.4 验证错误提示消息中提到的有效选项为 ":primary, :light, :vision"

### 3.3 文档验证
- [ ] 3.3.1 检查所有文档中的代码示例是否使用新的命名
- [ ] 3.3.2 验证配置文件示例中的字段名是否更新
- [ ] 3.3.3 验证命令行示例中的参数是否更新

## 4. 清理和收尾

### 4.1 代码清理
- [x] 4.1.1 使用全局搜索确认没有遗留的 `main`/`fast` 引用（排除 vision 相关和无关的 main 关键字）
- [x] 4.1.2 清理任何注释或文档字符串中的旧命名

### 4.2 提交前检查
- [x] 4.2.1 运行完整的测试套件：`go test ./...`
- [x] 4.2.2 运行静态分析：`go vet ./...`
- [x] 4.2.3 检查代码格式：`go fmt ./...`
- [x] 4.2.4 验证项目能正常编译和运行
