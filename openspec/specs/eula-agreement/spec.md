# EULA Agreement

## Purpose

定义 NFA 首次启动时的用户协议确认流程，确保用户在同意使用条款后方可使用软件，并支持协议版本变更后的重新确认机制。

## Requirements

### Requirement: EULA 协议内嵌

系统 SHALL 在 `pkg/eula/` 包中通过 Go embed 机制内嵌一份 `eula.md` 协议文件，该文件内容在编译时固定。

#### Scenario: 读取内嵌协议

- **WHEN** 程序需要展示或计算 EULA 内容时
- **THEN** 从 embed.FS 中读取内嵌的 `eula.md` 文件内容

### Requirement: 协议文件写入

每次启动时，系统 SHALL 将当前内嵌的 EULA 协议内容写入用户数据目录下的 `eula.md` 文件（`~/.nfa/eula.md`），使用 `0o644` 权限。

#### Scenario: 写入协议文件

- **WHEN** 程序启动且 `~/.nfa/` 目录已创建
- **THEN** 系统将内嵌的 `eula.md` 内容覆盖写入 `~/.nfa/eula.md`

### Requirement: 首次启动 EULA 确认

系统在首次启动（即 `~/.nfa/eula_signed.sha256sum` 文件不存在）时，SHALL 在终端展示 EULA 协议全文，并询问用户是否同意。

#### Scenario: 用户同意协议

- **WHEN** 首次启动展示 EULA 后用户输入同意（如 `y` 或 `yes`）
- **THEN** 系统计算当前内嵌 `eula.md` 的 SHA256 哈希值，写入 `~/.nfa/eula_signed.sha256sum`，并继续正常启动

#### Scenario: 用户拒绝协议

- **WHEN** 首次启动展示 EULA 后用户输入拒绝（如 `n` 或 `no`）
- **THEN** 系统输出提示信息并退出，退出码为非零

#### Scenario: 用户输入无效

- **WHEN** 展示 EULA 后用户输入既非同意也非拒绝的无效内容
- **THEN** 系统提示输入无效并要求重新输入

### Requirement: 已签署版本匹配时静默通过

系统 SHALL 在启动时检测 `~/.nfa/eula_signed.sha256sum` 文件。若该文件存在且其内容与当前内嵌 `eula.md` 的 SHA256 哈希值一致，则无需再次询问，静默继续启动。

#### Scenario: 签名匹配静默启动

- **WHEN** `~/.nfa/eula_signed.sha256sum` 存在且 SHA256 值与当前版本一致
- **THEN** 系统不展示任何 EULA 相关内容，直接继续启动

### Requirement: 协议变更时重新确认

系统 SHALL 在检测到 `~/.nfa/eula_signed.sha256sum` 存在但 SHA256 值与当前内嵌 `eula.md` 不匹配时，提示用户 EULA 已发生变更需要重新签署，展示最新协议内容并询问是否同意。

#### Scenario: 检测到协议变更

- **WHEN** `~/.nfa/eula_signed.sha256sum` 存在但 SHA256 值与当前版本不匹配
- **THEN** 系统输出"协议已更新"的提示，展示最新协议全文，并询问用户是否同意

#### Scenario: 同意更新后的协议

- **WHEN** 协议变更提示后用户输入同意
- **THEN** 系统更新 `~/.nfa/eula_signed.sha256sum` 为新的 SHA256 值，并继续正常启动

#### Scenario: 拒绝更新后的协议

- **WHEN** 协议变更提示后用户输入拒绝
- **THEN** 系统输出提示信息并退出，退出码为非零

### Requirement: EULA 仅对主命令生效

EULA 确认流程 SHALL 仅在用户执行 agent 主命令（根命令，无子命令）时触发。管理子命令（`version`、`models list`/`ls`、`internal-tools`、`otter`）MUST NOT 触发 EULA 检查。

#### Scenario: 执行主命令触发 EULA 检查

- **WHEN** 用户执行 `nfa` 根命令（带或不带 prompt 参数，带或不带 `-p` 标志）
- **THEN** 系统在启动前执行 EULA 检查流程

#### Scenario: 执行管理子命令跳过 EULA 检查

- **WHEN** 用户执行 `nfa version`、`nfa models list` 或 `nfa otter` 等子命令
- **THEN** 系统不执行 EULA 检查，直接执行子命令逻辑

### Requirement: 国际化支持

EULA 确认流程中的所有提示文本 SHALL 定义为 `i18n.Message` 结构体，支持中文和英文，与项目中其他模块的国际化模式保持一致。

#### Scenario: 中文环境提示

- **WHEN** 用户语言设置为中文（`--lang zh` 或配置中 `language: "zh"`）
- **THEN** EULA 确认提示以中文显示

#### Scenario: 英文环境提示

- **WHEN** 用户语言设置为英文（`--lang en` 或配置中 `language: "en"`）
- **THEN** EULA 确认提示以英文显示
