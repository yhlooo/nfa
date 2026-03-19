# 设计：一键安装脚本

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                      用户执行安装命令                        │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
┌──────────────────────────┐    ┌──────────────────────────┐
│      install.sh          │    │      install.ps1         │
│      (Linux/Darwin)      │    │      (Windows)           │
└──────────────────────────┘    └──────────────────────────┘
              │                               │
              └───────────────┬───────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   GitHub Releases API                       │
│         https://api.github.com/repos/yhlooo/nfa/           │
│                      releases/latest                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   下载对应平台资产                           │
│         nfa-v{version}-{os}-{arch}.tar.gz                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   解压并安装到目标目录                       │
└─────────────────────────────────────────────────────────────┘
```

## install.sh 详细设计

### 执行流程

```
┌─────────────────┐
│     开始        │
└────────┬────────┘
         ▼
┌─────────────────┐
│  检测 OS/ARCH   │  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
│                 │  ARCH=$(uname -m) → 转换为 amd64/arm64
└────────┬────────┘
         ▼
┌─────────────────┐
│  获取最新版本   │  curl -s https://api.github.com/repos/...
│                 │  grep '"tag_name"' 解析
└────────┬────────┘
         ▼
┌─────────────────┐
│  创建临时目录   │  mkdir -p ~/.cache/nfa
└────────┬────────┘
         ▼
┌─────────────────┐
│  下载压缩包     │  curl -L -o ~/.cache/nfa/nfa.tar.gz ...
└────────┬────────┘
         ▼
┌─────────────────┐
│  创建安装目录   │  mkdir -p ~/.local/bin
└────────┬────────┘
         ▼
┌─────────────────┐
│  解压安装       │  tar -xzf ... -C ~/.local/bin
└────────┬────────┘
         ▼
┌─────────────────┐
│  检测 PATH      │  $PATH 是否包含 ~/.local/bin
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌───────┐  ┌───────────────┐
│ 在PATH│  │ 不在PATH      │
│       │  │ 黄色提示设置  │
└───┬───┘  └───────┬───────┘
    └────┬─────────┘
         ▼
┌─────────────────┐
│  输出成功信息   │
└────────┬────────┘
         ▼
┌─────────────────┐
│     结束        │
└─────────────────┘
```

### ARCH 转换规则

| uname -m 输出 | 转换结果 |
|---------------|----------|
| x86_64 | amd64 |
| amd64 | amd64 |
| aarch64 | arm64 |
| arm64 | arm64 |

### PATH 提示逻辑

```bash
# 检测当前 shell
SHELL_NAME=$(basename "$SHELL")

# 判断 PATH 是否包含安装目录
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    # 黄色提示
    echo -e "\033[33m提示: ~/.local/bin 不在 PATH 中\033[0m"

    # 根据 shell 给出示例
    case "$SHELL_NAME" in
        bash)
            echo "请执行: echo 'export PATH=\"\${HOME}/.local/bin:\${PATH}\"' >> ~/.bashrc"
            echo "然后执行: source ~/.bashrc"
            ;;
        zsh)
            echo "请执行: echo 'export PATH=\"\${HOME}/.local/bin:\${PATH}\"' >> ~/.zshrc"
            echo "然后执行: source ~/.zshrc"
            ;;
        *)
            echo "请将 ~/.local/bin 添加到 PATH 环境变量"
            ;;
    esac
fi
```

## install.ps1 详细设计

### 执行流程

```
┌─────────────────┐
│     开始        │
└────────┬────────┘
         ▼
┌─────────────────┐
│  检测 ARCH      │  [Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")
│                 │  AMD64 → amd64, ARM64 → arm64
└────────┬────────┘
         ▼
┌─────────────────┐
│  获取最新版本   │  Invoke-RestMethod https://api.github.com/...
└────────┬────────┘
         ▼
┌─────────────────┐
│  创建临时目录   │  $env:TEMP\nfa
└────────┬────────┘
         ▼
┌─────────────────┐
│  下载压缩包     │  curl.exe -L -o ...
└────────┬────────┘
         ▼
┌─────────────────┐
│  创建安装目录   │  $env:LOCALAPPDATA\nfa
└────────┬────────┘
         ▼
┌─────────────────┐
│  解压安装       │  tar -xzf ...
└────────┬────────┘
         ▼
┌─────────────────┐
│  检测 PATH      │  $env:PATH 是否包含安装目录
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌───────┐  ┌───────────────┐
│ 在PATH│  │ 不在PATH      │
│       │  │ 黄色提示设置  │
└───┬───┘  └───────┬───────┘
    └────┬─────────┘
         ▼
┌─────────────────┐
│  输出成功信息   │
└────────┬────────┘
         ▼
┌─────────────────┐
│     结束        │
└─────────────────┘
```

### PATH 提示逻辑 (PowerShell)

```powershell
$InstallDir = "$env:LOCALAPPDATA\nfa"

# 检查 PATH 是否包含安装目录
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$InstallDir*") {
    Write-Host "提示: $InstallDir 不在 PATH 中" -ForegroundColor Yellow
    Write-Host "请执行以下命令添加到 PATH:"
    Write-Host '[Environment]::SetEnvironmentVariable("PATH", "$env:PATH;' + $InstallDir + '", "User")'
    Write-Host "然后重新打开终端窗口"
}
```

## 错误处理

两个脚本都采用极简策略：

- `set -e` (bash) / `$ErrorActionPreference = "Stop"` (PowerShell)
- 任何步骤失败直接退出并报错
- 不进行重试或 fallback

## 文件结构

```
scripts/
├── install.sh    # Linux/Darwin 安装脚本
└── install.ps1   # Windows 安装脚本
```
