#
# NFA 一键安装脚本 (Windows PowerShell)
# 用法: iex (irm https://raw.githubusercontent.com/yhlooo/nfa/refs/heads/master/scripts/install.ps1)
#

$ErrorActionPreference = "Stop"

$Repo = "yhlooo/nfa"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\nfa" }
$TmpDir = if ($env:TMP_DIR) { $env:TMP_DIR } else { "$env:TEMP\nfa" }

# 检测架构
function Get-Arch {
    $arch = [Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-Error "不支持的架构: $arch"
            exit 1
        }
    }
}

# 获取最新版本
function Get-LatestVersion {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    if (-not $release.tag_name) {
        Write-Error "无法获取最新版本"
        exit 1
    }
    return $release.tag_name
}

# 检查 PATH 是否包含安装目录
function Test-PathContains {
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    return $userPath -like "*$InstallDir*"
}

# 提示设置 PATH
function Write-PathPrompt {
    Write-Host ""
    Write-Host "提示: $InstallDir 不在 PATH 中" -ForegroundColor Yellow
    Write-Host "请执行以下命令添加到 PATH:"
    Write-Host "  [Environment]::SetEnvironmentVariable('PATH', `"`$env:PATH;$InstallDir`", 'User')"
    Write-Host "然后重新打开终端窗口"
}

# 主函数
function Main {
    Write-Host "正在安装 NFA..."

    # 检测系统信息
    $arch = Get-Arch
    Write-Host "检测到系统: windows/$arch"

    # 获取最新版本
    $version = Get-LatestVersion
    Write-Host "最新版本: $version"

    # 创建临时目录
    New-Item -ItemType Directory -Force -Path $TmpDir | Out-Null

    # 下载
    $assetName = "nfa-$version-windows-$arch.tar.gz"
    $downloadUrl = "https://github.com/$Repo/releases/download/$version/$assetName"
    $tmpFile = Join-Path $TmpDir $assetName

    Write-Host "正在下载 $assetName..."
    curl.exe -L -o $tmpFile $downloadUrl

    # 创建安装目录
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

    # 解压安装 (Windows 10+ 自带 tar)
    Write-Host "正在安装到 $InstallDir..."
    tar -xzf $tmpFile -C $InstallDir

    # 清理临时文件
    Remove-Item -Force $tmpFile

    # 检查 PATH
    if (-not (Test-PathContains)) {
        Write-PathPrompt
    }

    Write-Host ""
    Write-Host "安装成功!" -ForegroundColor Green
    Write-Host "安装位置: $InstallDir\nfa.exe"

    if (Test-PathContains) {
        Write-Host ""
        Write-Host "运行 'nfa --help' 查看使用帮助"
    }
}

Main
