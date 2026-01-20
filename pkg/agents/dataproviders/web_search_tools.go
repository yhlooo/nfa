package dataproviders

import (
	"time"

	"github.com/firebase/genkit/go/ai"
)

const (
	WebSearchToolName = "WebSearch"
	WebSearchDesc     = `Searching for specified keywords on the internet, returns a list of related websites,
each containing a website URL and a brief description.`
)

// WebSearchInput 网络搜索输入
type WebSearchInput struct {
	// 查询关键词
	Query string `json:"query"`
}

// WebSearchOutput 网络搜索输出
type WebSearchOutput struct {
	Items []WebSearchResultItem `json:"items"`
}

// WebSearchResultItem 搜索结果项
type WebSearchResultItem struct {
	// 标题
	Title string `json:"title"`
	// 简介
	Description string `json:"description"`
	// 发布日期
	Date time.Time `json:"date,omitempty"`
	// 网站 URL
	URL string `json:"url"`
	// 所属网站
	Site string `json:"site,omitempty"`
}

// WebSearchTool 网络搜索工具
type WebSearchTool = *ai.ToolDef[WebSearchInput, WebSearchOutput]
