package webbrowse

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// NewWebBrowser 创建 Web 浏览器
func NewWebBrowser() *WebBrowser {
	return &WebBrowser{}
}

// WebBrowser 网页浏览工具
//
// 支持 JS 运行、页面渲染，通过视觉模型读取网页内容。依赖 Chrome 和视觉模型
type WebBrowser struct {
	lock sync.Mutex

	cacheURL        string
	cacheScreenshot []byte
	cacheText       string
}

const (
	// BrowseToolName 网页浏览工具名
	BrowseToolName = "WebBrowse"
	// FetchToolName 网页内容获取工具名
	FetchToolName = "WebFetch"
)

// BrowseInput 网页浏览输入
type BrowseInput struct {
	// 需要浏览的 URL
	URL string `json:"url"`
	// 浏览网页时希望解答的问题
	Question string `json:"question,omitempty"`
}

// BrowseOutput 网页浏览输出
type BrowseOutput struct {
	// 网页文本内容
	Text string `json:"text,omitempty"`
	// 对问题的回答
	Answer string `json:"answer,omitempty"`
}

// RegisterTools 注册工具
func (wb *WebBrowser) RegisterTools(g *genkit.Genkit) []ai.ToolRef {
	// TODO: 检测有无视觉模型、有无浏览器
	return []ai.ToolRef{
		wb.DefineBrowseTool(g),
		wb.DefineFetchTool(g),
	}
}

// DefineBrowseTool 定义浏览网页工具
func (wb *WebBrowser) DefineBrowseTool(g *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(g, BrowseToolName, `浏览网页内容返回网页中文本内容或根据网页内容回答问题

以 JSON 格式输入：
- **url**: (string,required) 浏览的网页 URL
- **question**: (string) 针对网页内容的提问。工具可通过视觉方式浏览网页完整内容并回答该问题。该字段为空时工具返回网页中的文本内容（不含布局、图表等视觉元素信息），建议使用该字段获取页面中的细节信息。
`,
		func(ctx *ai.ToolContext, in BrowseInput) (BrowseOutput, error) {
			wb.lock.Lock()
			defer wb.lock.Unlock()

			chromeCtx, cancel := chromedp.NewContext(ctx)
			defer cancel()

			if in.URL == "" {
				return BrowseOutput{}, fmt.Errorf("url is required")
			}

			// 打开网页
			if wb.cacheURL != in.URL {
				wb.cacheText = ""
				wb.cacheScreenshot = nil
				if err := chromedp.Run(chromeCtx,
					chromedp.Navigate(in.URL),
					chromedp.Text("body", &wb.cacheText),
					chromedp.FullScreenshot(&wb.cacheScreenshot, 50),
				); err != nil {
					return BrowseOutput{}, err
				}
			}

			// 没有问题，直接返回文本内容
			if in.Question == "" {
				return BrowseOutput{Text: wb.cacheText}, nil
			}

			// 使用视觉模型读取图片内容
			m, _ := ctxutil.ModelsFromContext(ctx)
			resp, err := genkit.Generate(ctx, g,
				ai.WithModelName(m.GetVision()),
				ai.WithMessages(
					ai.NewUserMessage(
						ai.NewMediaPart(
							"image/png",
							"data:image/png;base64,"+base64.StdEncoding.EncodeToString(wb.cacheScreenshot)),
						ai.NewTextPart("根据图片中的信息回答：\n"+in.Question),
					),
				),
			)
			if err != nil {
				return BrowseOutput{}, err
			}
			return BrowseOutput{Answer: resp.Text()}, nil
		},
	)
}

// FetchInput 获取 URL 内容输入
type FetchInput struct {
	URL string `json:"url"`
}

// FetchOutput 获取 URL 内容输出
type FetchOutput struct {
	StatusCode int    `json:"statusCode"`
	Content    string `json:"content"`
}

// DefineFetchTool 定义获取 URL 内容工具
func (wb *WebBrowser) DefineFetchTool(g *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(g, FetchToolName,
		`Fetch the specified URL, read its content, convert it into readable text, and output.

The following types of URLs are supported:
- HTTP (http:// or https://)
- File (file://)

The following types of content are supported:
- Plain Text
- HTML
- JSON
- PDF
`,
		func(ctx *ai.ToolContext, input FetchInput) (FetchOutput, error) {
			logger := logr.FromContextOrDiscard(ctx)

			resp, err := http.Get(input.URL)
			if err != nil {
				return FetchOutput{}, fmt.Errorf("send http request error: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			ret := FetchOutput{StatusCode: resp.StatusCode}

			var content []byte
			switch resp.Header.Get("Content-Type") {
			case "application/pdf":
				// PDF
				logger.Info("pdf content found")
				content, err = convertPDFToText(ctx, resp.Body)

			default:
				// 其它
				logger.Info(fmt.Sprintf("content type: %s", resp.Header.Get("Content-Type")))
				content, err = io.ReadAll(io.LimitReader(resp.Body, 100<<20))
			}
			if err != nil {
				return ret, fmt.Errorf("read response body error: %w", err)
			}

			ret.Content = string(content)

			return ret, nil
		},
	)
}

// convertPDFToText 将 PDF 转换为文本
func convertPDFToText(ctx context.Context, r io.Reader) ([]byte, error) {
	tmpfile, err := os.CreateTemp("", "nfa-pdf-")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	_, err = io.Copy(tmpfile, r)
	_ = tmpfile.Close()
	if err != nil {
		return nil, fmt.Errorf("read pdf error: %w", err)
	}

	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", tmpfile.Name(), "-")
	out := &bytes.Buffer{}
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("convert pdf to text error: %w", err)
	}

	_ = os.Remove(tmpfile.Name())

	return out.Bytes(), nil
}
