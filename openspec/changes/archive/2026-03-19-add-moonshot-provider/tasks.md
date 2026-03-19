# 任务清单

## 实现任务

- [x] **T1: 创建 Moonshot 提供商实现文件**
  - 创建 `pkg/models/moonshot.go`
  - 定义常量 `MoonshotProviderName` 和 `MoonshotBaseURL`
  - 实现 `MoonshotModels()` 返回建议模型列表
  - 实现 `MoonshotOptions` 结构体
  - 实现 `Complete()` 方法
  - 实现 `Plugin()` 方法
  - 实现 `RegisterModels()` 方法

- [x] **T2: 更新 ModelProvider 结构体**
  - 在 `pkg/models/providers.go` 中添加 `Moonshot *MoonshotOptions` 字段

- [x] **T3: 添加 i18n 消息定义**
  - 在 `pkg/models/i18n.go` 中添加 `MsgModelDescKimiK25`
  - 在 `pkg/i18n/active.zh.yaml` 中添加中文翻译
  - 在 `pkg/i18n/active.en.yaml` 中添加英文翻译

- [x] **T4: 更新 Genkit 初始化逻辑**
  - 在 `pkg/agents/genkit.go` 的 `NewGenkitWithModels` 函数中添加 Moonshot case

- [x] **T5: 更新配置文档**
  - 在 `docs/reference/config.md` 中添加 Moonshot 配置说明

## 验证任务

- [x] **T6: 代码质量检查**
  - 执行 `go fmt ./...`
  - 执行 `go vet ./...`
  - 执行 `go test ./...`

- [x] **T7: 功能验证**
  - 验证配置解析正确
  - 验证模型注册正确
  - 验证模型列表显示正确
