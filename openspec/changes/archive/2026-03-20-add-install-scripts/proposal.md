# 提案：添加一键安装脚本

## 概述

为 NFA 项目添加一键安装脚本，支持用户通过 curl/PowerShell 快速安装最新版本。

## 动机

目前用户需要手动从 GitHub Releases 下载对应平台的二进制文件并配置环境变量，安装门槛较高。提供一键安装脚本可以：

1. 降低用户安装门槛
2. 自动检测系统平台
3. 自动获取最新版本
4. 提供友好的安装提示

## 范围

### 包含

- `scripts/install.sh` - Linux/Darwin 安装脚本
- `scripts/install.ps1` - Windows 安装脚本

### 不包含

- 包管理器支持 (brew, winget, scoop 等)
- 版本选择安装
- Checksum 校验
- 自动配置 PATH 环境变量

## 设计要点

### install.sh (Linux/Darwin)

| 项目 | 值 |
|------|-----|
| 安装目录 | `~/.local/bin` |
| 临时目录 | `~/.cache/nfa` |
| 下载工具 | curl (wget 兜底) |
| 解压工具 | tar |
| PATH 提示 | 黄色，bash/zsh 示例 |

### install.ps1 (Windows)

| 项目 | 值 |
|------|-----|
| 安装目录 | `$env:LOCALAPPDATA\nfa` |
| 临时目录 | `$env:TEMP\nfa` |
| 下载工具 | curl.exe (Windows 自带) |
| 解压工具 | tar (Windows 10+ 自带) |
| PATH 提示 | 黄色，PowerShell 示例 |

## 使用方式

```bash
# Linux/Darwin
curl -L https://raw.githubusercontent.com/yhlooo/nfa/refs/heads/master/scripts/install.sh | bash

# Windows PowerShell
iex (irm https://raw.githubusercontent.com/yhlooo/nfa/refs/heads/master/scripts/install.ps1)
```

## 风险

| 风险 | 缓解措施 |
|------|----------|
| GitHub API 限流 | 使用未认证请求，每小时 60 次足够安装使用 |
| 网络问题 | 失败时报错退出，提示用户重试 |
| 权限问题 | 安装到用户目录，避免权限问题 |

## 成功标准

1. 执行 `curl -L ... | bash` 可在 Linux/Darwin 上成功安装
2. 执行 `iex (irm ...)` 可在 Windows 上成功安装
3. 安装后可执行 `nfa --help`
4. PATH 未配置时显示黄色提示
