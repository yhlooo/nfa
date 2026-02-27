# models list 命令输出优化 - 实现任务

## 1. 基础结构和命令行选项

- [x] 1.1 在 `pkg/commands/models.go` 中添加 `ModelsListOptions` 结构体，包含 Provider、Capability、Format 字段
- [x] 1.2 在 `newModelsListCommand` 中添加命令行标志：`--provider`、`--capability`、`--format`
- [x] 1.3 添加命令帮助文本和使用示例

## 2. 过滤逻辑实现

- [x] 2.1 实现 `filterModels` 函数，支持按提供商过滤
- [x] 2.2 实现 `filterModels` 函数，支持按能力过滤（reasoning/vision）
- [x] 2.3 支持组合过滤条件（provider + capability）
- [x] 2.4 添加无效能力名称的错误处理

## 3. 颜色和样式输出

- [x] 3.1 使用 `termenv` 创建输出样式函数 `newOutputStyles`
- [x] 3.2 定义颜色样式：提供商（青色）、推理（紫色）、视觉（蓝色）、默认标签（绿色加粗）、上下文（淡色）
- [x] 3.3 实现终端能力检测（NO_COLOR、非终端输出）

## 4. 文本格式化

- [x] 4.1 实现 `extractProvider` 函数，从模型名称提取提供商
- [x] 4.2 实现 `formatContextWindow` 函数，格式化上下文大小（128000 → 128K）
- [x] 4.3 实现 `renderModelLine` 函数，渲染单个模型的文本行
- [x] 4.4 使用 `github.com/mattn/go-runewidth` 实现列对齐逻辑：
  - 列 1: 模型名称（35字符，左对齐）
  - 列 2: Emoji+标签（12字符，左对齐）
  - 列 3: 上下文（8字符，右对齐）
  - 列 4: 描述（40字符，左对齐）
- [x] 4.5 实现 `truncateByWidth` 函数，按显示宽度截断字符串（使用 `runewidth.RuneWidth`）
- [x] 4.6 处理长模型名称的换行逻辑（超过 35 字符时描述换行）
- [x] 4.7 实现描述截断逻辑（超过 40 字符添加 "..."）

## 5. 输出函数实现

- [x] 5.1 实现 `outputTextModels` 函数，输出人性化文本格式
- [x] 5.2 添加 emoji 渲染：🧠（推理）、👁️（视觉）
- [x] 5.3 实现默认模型标签：[main]、[fast]、[vision]
- [x] 5.4 实现无匹配结果的提示信息

## 6. JSON 输出实现

- [x] 6.1 定义 JSON 输出结构体（ModelJSONOutput、ModelJSON）
- [x] 6.2 实现 `outputJSONModels` 函数，输出 JSON 格式
- [x] 6.3 实现模型的 tags 字段逻辑（main/fast/vision）
- [x] 6.4 确保输出包含所有模型元数据（name、provider、description、capabilities、contextWindow、maxOutputTokens、tags、cost）

## 7. 主命令逻辑

- [x] 7.1 修改 `newModelsListCommand` 的 RunE 函数，调用过滤和输出函数
- [x] 7.2 根据格式选项调用对应的输出函数（text 或 json）
- [x] 7.3 确保从配置中获取默认模型信息

## 8. 测试

- [x] 8.1 手动测试基本列表输出
- [x] 8.2 测试 emoji 显示宽度计算（🧠 和 👁️）
- [x] 8.3 测试长模型名称的截断和换行
- [x] 8.4 测试长描述的截断
- [x] 8.5 测试 `--provider` 过滤
- [x] 8.6 测试 `--capability` 过滤
- [x] 8.7 测试组合过滤
- [x] 8.8 测试 `--format json` 输出
- [x] 8.9 测试颜色在不同终端的表现
- [x] 8.10 测试 `NO_COLOR` 环境变量
- [x] 8.11 测试输出重定向到文件

## 9. 代码质量

- [x] 9.1 运行 `go fmt ./...` 格式化代码
- [x] 9.2 运行 `go vet ./...` 检查代码问题
- [x] 9.3 运行 `go test ./...` 确保测试通过
