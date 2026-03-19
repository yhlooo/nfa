**[简体中文](README_CN.md)** | [English](README.md)

---

![GitHub License](https://img.shields.io/github/license/yhlooo/nfa)
[![GitHub Release](https://img.shields.io/github/v/release/yhlooo/nfa)](https://github.com/yhlooo/nfa/releases/latest)
[![release](https://github.com/yhlooo/nfa/actions/workflows/release.yaml/badge.svg)](https://github.com/yhlooo/nfa/actions/workflows/release.yaml)

# NFA (Not Financial Advice, 非财务建议)

基于大语言模型的金融交易顾问 Agent 。

> **注意：该程序任何输出都不应被理解为财务建议。**

## 安装

### 通过二进制安装

#### Linux / macOS

```bash
curl -L https://raw.githubusercontent.com/yhlooo/nfa/refs/heads/master/scripts/install.sh | bash
```

脚本将 `nfa` 安装到 `~/.local/bin`。如果该目录不在 `PATH` 中，请按照提示添加。

#### Windows

```powershell
iex (irm https://raw.githubusercontent.com/yhlooo/nfa/refs/heads/master/scripts/install.ps1)
```

脚本将 `nfa` 安装到 `$env:LOCALAPPDATA\nfa`。如果该目录不在 `PATH` 中，请按照提示添加。

#### 手动安装

通过 [Releases](https://github.com/yhlooo/nfa/releases) 页面下载可执行二进制，解压并将其中 `nfa` 文件放置到任意 `$PATH` 目录下。

### Docker

使用镜像 [`ghcr.io/yhlooo/nfa`](https://github.com/yhlooo/nfa/pkgs/container/nfa) 直接 docker run：

```bash
docker run -v "${HOME}/.nfa:/root/.nfa" -it --rm ghcr.io/yhlooo/nfa:latest --help
```

### 从源码编译

要求 Go 1.24.7 或更高版本，执行以下命令下载源码并构建：

```bash
go install github.com/yhlooo/nfa/cmd/nfa@latest
```

构建的二进制默认将在 `${GOPATH}/bin` 目录下，需要确保该目录包含在 `$PATH` 中。

## 使用

在 `~/.nfa/nfa.json` 配置模型、数据源，见 [配置参考](docs/reference/config.md)

然后启动交互式聊天会话：

```bash
nfa
```

或单次回答后退出：

```bash
nfa -p "什么是市盈率？"
```

### 模型选择

在对话中切换模型：

```
/model                             # 交互式选择切换主模型
/model deepseek/deepseek-reasoner  # 直接切换

/model :vision  # 切换视觉理解模型
```

### 自定义技能

NFA 支持添加自定义技能扩展 Agent 能力，在 `~/.nfa/skills/<技能名>/SKILL.md` 中创建自定义技能：

```markdown
---
name: get-price
description: 获取资产价格
---

1. 首先确认资产正确代码
2. 查询资产近 5 个交易日的价格
3. 返回价格数据，包括日期和收盘价
```

Agent 会自动加载并在需要时使用这些技能。
