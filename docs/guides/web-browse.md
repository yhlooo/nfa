# 网页浏览 (Web Browse)

NFA 支持通过网页浏览工具读取网页内容，实现对互联网信息的深度访问和理解。

## 概述

NFA 提供两个网页浏览工具：

- **WebBrowse** - 使用 Chrome 浏览器加载网页，支持 JavaScript 渲染，可通过视觉模型理解页面内容
- **WebFetch** - 直接获取 URL 内容并转换为可读文本，支持多种文件格式

这两个工具相辅相成，Agent 会根据需要选择合适的工具。

## WebBrowse 工具

### 功能特点

WebBrowse 工具提供了完整的浏览器能力：

- **JavaScript 执行** - 可以渲染动态生成的内容
- **视觉理解** - 通过视觉模型理解页面的布局、图表等视觉元素
- **智能问答** - 可以针对网页内容提出问题，获得针对性的回答
- **缓存机制** - 同一 URL 只加载一次，提高效率

### 依赖条件

使用 WebBrowse 需要满足以下条件：

1. **Chrome 浏览器** - 系统需要安装 Chrome 浏览器
2. **视觉模型** - 需要配置支持视觉理解的模型

### 配置视觉模型

在 `~/.nfa/nfa.json` 中配置视觉模型：

```json
{
  "defaultModels": {
    "main": "ollama/llama2",
    "fast": "ollama/llama2",
    "vision": "openaiCompatible/gpt-4-vision-preview"
  }
}
```

注意：如果 `vision` 字段留空或配置的模型不支持视觉，WebBrowse 的问答功能将无法使用，但仍可以返回文本内容。

### 输入输出格式

**输入**:
```json
{
  "url": "https://example.com",
  "question": "这个页面的主要内容是什么？"
}
```

**输出**:
- 如果没有 `question`，返回网页的纯文本内容
- 如果有 `question`，返回视觉模型对页面内容的回答

**输出结构**:
```json
{
  "text": "网页的纯文本内容",
  "answer": "对问题的回答"
}
```

### 使用场景

#### 场景 1：读取新闻网站

当 Agent 需要阅读具体的新闻文章时：

```
用户：详细分析一下最新的特斯拉财报
```

Agent 可能会：
1. 搜索找到财报页面
2. 使用 WebBrowse 读取财报内容
3. 针对具体指标（营收、利润、现金流）提问

#### 场景 2：理解复杂页面

对于包含图表、布局复杂的页面：

```
用户：这个网页上的折线图显示了什么趋势？
```

视觉模型可以识别和理解图表内容，并用自然语言描述趋势。

## WebFetch 工具

### 功能特点

WebFetch 是一个轻量级的网页内容获取工具：

- **无需浏览器** - 直接通过 HTTP 请求获取内容
- **多格式支持** - 自动将 HTML、PDF、JSON 等格式转换为可读文本
- **本地文件支持** - 通过 `file://` 协议读取本地文件

### 依赖条件

WebFetch 需要以下额外工具：

- **pdftotext** - 用于处理 PDF 文件（来自 poppler-utils）

**安装方法**:
```bash
# Ubuntu/Debian
sudo apt install poppler-utils

# macOS
brew install poppler
```

### 输入输出格式

**输入**:
```json
{
  "url": "https://example.com/document.pdf"
}
```

**输出**:
```json
{
  "statusCode": 200,
  "content": "转换后的文本内容"
}
```

### 支持的格式

| 协议 | 内容类型 |
|------|----------|
| `http://` / `https://` | HTML、纯文本、JSON、PDF |
| `file://` | 本地 HTML、纯文本、JSON、PDF |

### 使用场景

#### 场景 1：读取 API 文档

```
用户：帮我查看这个 API 的使用方法
```

Agent 可以使用 WebFetch 快速获取 API 文档内容。

#### 场景 2：读取本地文件

```
用户：帮我分析一下本地这个文件的 content
```

Agent 可以通过 `file://` 协议读取本地文件。

## 工具选择逻辑

Agent 会根据任务需求自动选择合适的工具：

| 情况 | 推荐工具 | 原因 |
|------|----------|------|
| 需要理解页面布局、图表 | WebBrowse | 需要视觉能力 |
| 网页有大量 JavaScript | WebBrowse | 需要渲染执行 |
| 简单的文本内容 | WebFetch | 更快速高效 |
| 读取 PDF 文件 | WebFetch | 自动转换 |
| 读取本地文件 | WebFetch | 支持 file:// |

## 注意事项

1. **性能考虑** - WebBrowse 需要启动浏览器和截图，比 WebFetch 慢，只在实际需要时使用

2. **缓存机制** - WebBrowse 对同一 URL 有缓存，避免重复加载

3. **错误处理** - 如果网页加载失败或视觉模型不可用，Agent 会记录错误并尝试其他方法

4. **内容大小限制** - WebFetch 限制最大读取 100MB 内容，超大文件可能无法完整处理

5. **权限问题** - 某些网站可能有反爬虫机制，或需要登录才能访问，这种情况下 WebBrowse 和 WebFetch 都可能失败

## 配置示例

完整配置示例（包含模型提供商和视觉模型）：

```json
{
  "modelProviders": [
    {
      "ollama": {
        "serverAddress": "http://localhost:11434",
        "timeout": 300
      }
    },
    {
      "openaiCompatible": {
        "name": "openai",
        "baseURL": "https://api.openai.com/v1",
        "apiKey": "your-openai-api-key"
      }
    }
  ],
  "defaultModels": {
    "main": "ollama/llama2",
    "fast": "ollama/llama2",
    "vision": "openai/gpt-4-vision-preview"
  },
  "dataProviders": [...]
}
```

## 故障排查

### WebBrowse 无法使用

1. **检查 Chrome 是否安装**
   ```bash
   google-chrome --version  # Linux
   /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --version  # macOS
   ```

2. **检查视觉模型是否配置**
   - 确保 `defaultModels.vision` 字段已填写
   - 确认配置的模型支持视觉功能

3. **查看日志**
   - 检查 `~/.nfa/nfa.log` 中的错误信息

### PDF 文件无法读取

1. **检查 pdftotext 是否安装**
   ```bash
   pdftotext -v
   ```

2. **检查 PDF 文件大小**
   - 超过 100MB 的文件可能无法完整读取

3. **检查 PDF 文件是否损坏**
   - 尝试用其他 PDF 阅读器打开验证

### 网页加载失败

1. **检查网络连接**
   - 确保可以访问目标网站

2. **检查网站限制**
   - 某些网站有地区限制或反爬虫保护

3. **检查 URL 格式**
   - 确保使用正确的协议（http:// 或 https://）
