#!/bin/bash
#
# NFA 一键安装脚本
# 用法: curl -L https://raw.githubusercontent.com/yhlooo/nfa/refs/heads/master/scripts/install.sh | bash
#

set -e

REPO="yhlooo/nfa"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TMP_DIR="${TMP_DIR:-$HOME/.cache/nfa}"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 检测操作系统
detect_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux|darwin)
            echo "$os"
            ;;
        *)
            echo "不支持的操作系统: $os" >&2
            exit 1
            ;;
    esac
}

# 检测架构
detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            echo "不支持的架构: $arch" >&2
            exit 1
            ;;
    esac
}

# 获取最新版本
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        echo "无法获取最新版本" >&2
        exit 1
    fi
    echo "$version"
}

# 下载文件
download() {
    local url="$1"
    local output="$2"

    if command -v curl &> /dev/null; then
        curl -L -o "$output" "$url"
    elif command -v wget &> /dev/null; then
        wget -O "$output" "$url"
    else
        echo "需要 curl 或 wget" >&2
        exit 1
    fi
}

# 检查 PATH 是否包含安装目录
check_path() {
    case ":$PATH:" in
        *":$INSTALL_DIR:"*)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# 提示设置 PATH
prompt_path() {
    local shell_name
    shell_name=$(basename "$SHELL" 2>/dev/null || echo "unknown")

    echo ""
    echo -e "${YELLOW}提示: $INSTALL_DIR 不在 PATH 中${NC}"

    case "$shell_name" in
        bash)
            echo "请执行以下命令添加到 PATH:"
            echo "  echo 'export PATH=\"\${HOME}/.local/bin:\${PATH}\"' >> ~/.bashrc"
            echo "  source ~/.bashrc"
            ;;
        zsh)
            echo "请执行以下命令添加到 PATH:"
            echo "  echo 'export PATH=\"\${HOME}/.local/bin:\${PATH}\"' >> ~/.zshrc"
            echo "  source ~/.zshrc"
            ;;
        *)
            echo "请将 $INSTALL_DIR 添加到 PATH 环境变量"
            ;;
    esac
}

main() {
    echo "正在安装 NFA..."

    # 检测系统信息
    local os arch version
    os=$(detect_os)
    arch=$(detect_arch)
    echo "检测到系统: $os/$arch"

    # 获取最新版本
    version=$(get_latest_version)
    echo "最新版本: $version"

    # 创建临时目录
    mkdir -p "$TMP_DIR"

    # 下载
    local asset_name="nfa-${version}-${os}-${arch}.tar.gz"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${asset_name}"
    local tmp_file="${TMP_DIR}/${asset_name}"

    echo "正在下载 $asset_name..."
    download "$download_url" "$tmp_file"

    # 创建安装目录
    mkdir -p "$INSTALL_DIR"

    # 解压安装
    echo "正在安装到 $INSTALL_DIR..."
    tar -xzf "$tmp_file" -C "$INSTALL_DIR"

    # 设置执行权限
    chmod +x "${INSTALL_DIR}/nfa"

    # 清理临时文件
    rm -f "$tmp_file"

    # 检查 PATH
    if ! check_path; then
        prompt_path
    fi

    echo ""
    echo -e "${GREEN}安装成功!${NC}"
    echo "安装位置: ${INSTALL_DIR}/nfa"

    if check_path; then
        echo ""
        echo "运行 'nfa --help' 查看使用帮助"
    fi
}

main "$@"
