## 1. 数据结构扩展

- [x] 1.1 在 `SkillMeta` 结构体中添加 `Source` 字段（`yaml:"-"`，避免从 YAML 解析）
- [x] 1.2 在 `SkillRef` 结构体中添加 `Source` 字段用于存储技能来源
- [x] 1.3 定义技能来源常量 `SkillSourceBuiltin = "builtin"` 和 `SkillSourceLocal = "local"`

## 2. 内置技能文件结构

- [x] 2.1 创建 `pkg/skills/builtin/` 目录
- [x] 2.2 创建 `pkg/skills/builtin/short-term-trend-forecast/SKILL.md` 文件
- [x] 2.3 将现有 `~/.nfa/skills/short-term-trend-forecast/SKILL.md` 内容复制到内置技能文件

## 3. embed 文件系统实现

- [x] 3.1 创建 `pkg/skills/builtin.go` 文件
- [x] 3.2 在 `builtin.go` 中声明 `//go:embed builtin` 指令
- [x] 3.3 声明 `var builtinSkillFS embed.FS` 变量
- [x] 3.4 实现 `loadBuiltinSkills(ctx context.Context)` 方法，从 embed.FS 读取内置技能

## 4. 技能加载逻辑修改

- [x] 4.1 修改 `LoadMeta()` 方法，先调用 `loadBuiltinSkills()` 加载内置技能
- [x] 4.2 修改 `LoadMeta()` 方法，后调用现有逻辑加载用户技能（用户技能覆盖同名内置技能）
- [x] 4.3 修改 `loadUserSkills()` 重构现有逻辑，将用户技能来源标记为 `local`
- [x] 4.4 修改 `Get()` 方法，根据 `SkillRef.Source` 判断从 embed.FS 或文件系统读取

## 5. 工具描述更新

- [x] 5.1 更新 `DefineSkillTool()` 中的工具描述文本，说明技能可来自内置或用户定义位置
- [x] 5.2 确保工具描述提及内置技能（embed）和用户技能（`~/.nfa/skills/`）两个来源

## 6. 单元测试

- [x] 6.1 添加测试用例：仅内置技能可用时的加载行为
- [x] 6.2 添加测试用例：内置技能和用户技能共存时的合并行为
- [x] 6.3 添加测试用例：用户技能覆盖同名内置技能的行为
- [x] 6.4 添加测试用例：从 embed.FS 读取内置技能内容
- [x] 6.5 添加测试用例：`SkillMeta.Source` 字段正确标记来源
- [x] 6.6 添加测试用例：内置技能解析失败时的错误处理

## 7. 集成测试

- [x] 7.1 运行现有集成测试，确保向后兼容
- [x] 7.2 添加集成测试：Agent 调用内置技能的完整流程
- [x] 7.3 添加集成测试：Agent 调用用户技能的完整流程

## 8. 代码质量检查

- [x] 8.1 运行 `go fmt ./...` 格式化代码
- [x] 8.2 运行 `go vet ./...` 检查代码问题
- [x] 8.3 运行 `go test ./...` 确保所有测试通过

## 9. 文档更新（可选）

- [x] 9.1 更新 `docs/guides/` 中关于技能系统的使用指南（如有必要）
- [x] 9.2 添加内置技能开发说明到相关文档（如有必要）
