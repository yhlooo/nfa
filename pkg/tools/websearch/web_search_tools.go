package websearch

import (
	"time"

	"github.com/firebase/genkit/go/ai"
)

const (
	SearchToolName = "WebSearch"
	SearchDesc     = `Searching for specified keywords on the internet, returns a list of related websites,
each containing a website URL and a brief description.`
)

// SearchInput 网络搜索输入
type SearchInput struct {
	// 查询关键词
	Query string `json:"query"`
}

// SearchOutput 网络搜索输出
type SearchOutput struct {
	Items []SearchResultItem `json:"items"`
}

// SearchResultItem 搜索结果项
type SearchResultItem struct {
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

// SearchTool 网络搜索工具
type SearchTool = *ai.ToolDef[SearchInput, SearchOutput]
