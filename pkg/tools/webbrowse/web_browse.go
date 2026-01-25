package webbrowse

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

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

// WebBrowsingInput 网页浏览输入
type WebBrowsingInput struct {
	// 需要浏览的 URL
	URL string `json:"url"`
	// 浏览网页时希望解答的问题
	Question string `json:"question,omitempty"`
}

// WebBrowsingOutput 网页浏览输出
type WebBrowsingOutput struct {
	// 网页文本内容
	Text string `json:"text,omitempty"`
	// 对问题的回答
	Answer string `json:"answer,omitempty"`
}

const (
	// WebBrowseToolName 网页浏览工具名
	WebBrowseToolName = "WebBrowse"
)

// DefineBrowseTool 定义浏览网页工具
func (wb *WebBrowser) DefineBrowseTool(g *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(g, WebBrowseToolName, `浏览网页内容返回网页中文本内容或根据网页内容回答问题

以 JSON 格式输入：
- **url**: (string,required) 浏览的网页 URL
- **question**: (string) 针对网页内容的提问。工具可通过视觉方式浏览网页完整内容并回答该问题。该字段为空时工具返回网页中的文本内容（不含布局、图表等视觉元素信息），建议使用该字段获取页面中的细节信息。
`,
		func(ctx *ai.ToolContext, in WebBrowsingInput) (WebBrowsingOutput, error) {
			wb.lock.Lock()
			defer wb.lock.Unlock()

			chromeCtx, cancel := chromedp.NewContext(ctx)
			defer cancel()

			if in.URL == "" {
				return WebBrowsingOutput{}, fmt.Errorf("url is required")
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
					return WebBrowsingOutput{}, err
				}
			}

			// 没有问题，直接返回文本内容
			if in.Question == "" {
				return WebBrowsingOutput{Text: wb.cacheText}, nil
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
				return WebBrowsingOutput{}, err
			}
			return WebBrowsingOutput{Answer: resp.Text()}, nil
		},
	)
}
