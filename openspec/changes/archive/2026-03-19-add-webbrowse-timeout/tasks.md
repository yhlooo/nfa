# Tasks: WebBrowse 工具超时功能

## 任务列表

- [x] **T1: 修改 BrowseInput 结构体**
  - 文件：`pkg/tools/webbrowse/web_browse.go`
  - 添加 `Timeout` 字段（int，json tag 为 `timeout`）
  - 更新工具描述文档，说明 timeout 参数

- [x] **T2: 实现超时逻辑**
  - 文件：`pkg/tools/webbrowse/web_browse.go`
  - 在 `DefineBrowseTool` 函数入口处添加 `context.WithTimeout`
  - 默认值 60 秒
  - 将超时 context 传递给 chromedp 和 genkit 操作

- [x] **T3: 更新文档**
  - 文件：`docs/guides/web-browse.md`
  - 在输入输出格式部分添加 `timeout` 参数说明
  - 添加超时相关注意事项

- [x] **T4: 代码质量检查**
  - 运行 `go fmt ./...`
  - 运行 `go vet ./...`
  - 运行 `go test ./...`

## 依赖关系

```
T1 ──▶ T2 ──▶ T4
         │
         └──▶ T3
```

T1 和 T2 可合并为一个任务，T3 可并行进行。
