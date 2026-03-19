# 任务：一键安装脚本

## 任务列表

- [x] 创建 `scripts/install.sh` Linux/Darwin 安装脚本
- [x] 创建 `scripts/install.ps1` Windows 安装脚本

## 验收标准

1. Linux/Darwin 上执行 `curl -L .../install.sh | bash` 可成功安装
2. Windows 上执行 `iex (irm .../install.ps1)` 可成功安装
3. 安装后执行 `nfa --help` 显示帮助信息
4. PATH 未包含安装目录时显示黄色提示
