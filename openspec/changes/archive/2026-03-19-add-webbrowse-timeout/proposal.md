# Proposal: WebBrowse 工具支持设置超时时间

## 问题陈述

当前 WebBrowse 工具没有超时设置，在某些情况下可能导致工具长时间阻塞：

- 网页加载缓慢
- 网络连接不稳定
- 页面资源过大
- 视觉模型响应慢

Agent 无法控制工具的执行时间，可能导致整个请求超时或资源浪费。

## 提议方案

为 WebBrowse 工具添加 `timeout` 参数：

- 参数类型：整数（秒）
- 默认值：60 秒
- 作用范围：整个工具执行过程（页面加载 + 视觉模型理解）
- 行为：超时后返回错误

## 影响范围

- `pkg/tools/webbrowse/web_browse.go` - 核心实现
- `docs/guides/web-browse.md` - 文档更新

## 非目标

- 不修改 WebFetch 工具
- 不添加全局配置项
