## Why

当前项目的使用说明文档直接放在 README.md 中，随着功能增多会导致 README 过于冗长，难以维护。同时缺少结构化的文档组织方式，不利于用户快速定位和查阅相关内容。

## What Changes

- 创建 `docs/` 目录，建立三层文档结构
  - `docs/tutorials/`: 分步教程，引导用户从入门到使用
  - `docs/guides/`: 按特性组织的详细使用说明
  - `docs/reference/`: API 和配置文件的参考信息
- 将 README.md 中的自定义技能说明迁移到 `docs/guides/skills.md`
- 重构 README.md，使其仅作为项目整体介绍和快速开始
- 在 README.md 中添加文档链接导航

## Capabilities

### New Capabilities
- `docs-structure`: 建立项目的文档目录结构，包含 tutorials、guides、reference 三个子目录

### Modified Capabilities
- (无)

## Impact

- 文件变更：创建 `docs/` 目录结构，修改 `README.md`
- 用户影响：文档位置变更，需要更新用户阅读路径
- 零代码变更，仅文档重组
