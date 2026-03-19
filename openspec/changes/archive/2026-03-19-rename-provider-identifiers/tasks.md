## 任务列表

### T1: 重命名代码文件
- [x] `git mv pkg/models/aliyun_dashscope.go pkg/models/qwen.go`
- [x] `git mv pkg/models/zhipu_bigmodel.go pkg/models/zai.go`

### T2: 更新 qwen.go 内部命名
- [x] 重命名常量 `DashScopeProviderName` → `QwenProviderName`，值改为 `"qwen"`
- [x] 重命名常量 `DashScopeBaseURL` → `QwenBaseURL`
- [x] 重命名函数 `DashScopeModels` → `QwenModels`
- [x] 重命名类型 `DashScopeOptions` → `QwenOptions`
- [x] 更新所有内部引用

### T3: 更新 zai.go 内部命名
- [x] 重命名常量 `BigModelProviderName` → `ZAIProviderName`，值改为 `"z-ai"`
- [x] 重命名常量 `BigModelBaseURL` → `ZAIBaseURL`
- [x] 重命名函数 `BigModelModels` → `ZAIModels`
- [x] 重命名类型 `BigModelOptions` → `ZAIOptions`
- [x] 更新所有内部引用

### T4: 更新 providers.go
- [x] 字段 `Zhipu` → `ZAI`，JSON tag `"zhipu"` → `"z-ai"`
- [x] 字段 `Aliyun` → `Qwen`，JSON tag `"aliyun"` → `"qwen"`
- [x] 更新字段类型引用

### T5: 更新 genkit.go
- [x] 更新 case 分支：`p.Zhipu` → `p.ZAI`
- [x] 更新 case 分支：`p.Aliyun` → `p.Qwen`
- [x] 更新错误日志消息中的供应商名

### T6: 更新文档
- [x] `docs/guides/model-config.md` - 更新配置示例和模型前缀
- [x] `docs/guides/command-line.md` - 更新命令示例

### T7: 更新活跃 spec 文件
- [x] `openspec/specs/model-config/spec.md` - 更新 "aliyun" 为 "qwen"
- [x] `openspec/specs/model-selection/spec.md` - 更新模型前缀示例

### T8: 代码质量检查
- [x] `go fmt ./...`
- [x] `go vet ./...`
- [x] `go test ./...`
