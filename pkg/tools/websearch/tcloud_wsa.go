package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	wsa "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/wsa/v20250508"
)

// TencentCloudWSAOptions 腾讯云 WSA 选项
type TencentCloudWSAOptions struct {
	Endpoint  string `json:"endpoint,omitempty"`
	SecretID  string `json:"secretID"`
	SecretKey string `json:"secretKey"`
}

// RegisterTool 注册工具
func (opts *TencentCloudWSAOptions) RegisterTool(_ context.Context, g *genkit.Genkit) (SearchTool, error) {
	cred := common.NewCredential(opts.SecretID, opts.SecretKey)
	p := profile.NewClientProfile()
	if opts.Endpoint != "" {
		p.HttpProfile.Endpoint = opts.Endpoint
	}
	client, err := wsa.NewClient(cred, "", p)
	if err != nil {
		return nil, fmt.Errorf("new tencent cloud wsa client error: %s", err)
	}

	return DefineTencentCloudWSASearchTool(g, client), nil
}

// DefineTencentCloudWSASearchTool 定义腾讯云 WSA 搜索工具
func DefineTencentCloudWSASearchTool(
	g *genkit.Genkit,
	client *wsa.Client,
) SearchTool {
	return genkit.DefineTool(g, SearchToolName, SearchDesc,
		func(ctx *ai.ToolContext, in SearchInput) (SearchOutput, error) {
			req := wsa.NewSearchProRequest()
			req.Query = &in.Query
			resp, err := client.SearchProWithContext(ctx, req)
			if err != nil {
				return SearchOutput{}, err
			}
			if resp == nil || resp.Response == nil {
				return SearchOutput{}, nil
			}
			ret := &SearchOutput{}
			for _, page := range resp.Response.Pages {
				if page == nil {
					continue
				}
				data := wsaPage{}
				if err := json.Unmarshal([]byte(*page), &data); err != nil {
					ret.Items = append(ret.Items, SearchResultItem{
						Description: *page,
					})
					continue
				}
				date, _ := time.Parse(time.DateTime, data.Date)
				ret.Items = append(ret.Items, SearchResultItem{
					Title:       data.Title,
					Description: data.Passage,
					Date:        date,
					URL:         data.URL,
					Site:        data.Site,
				})
			}
			return *ret, nil
		},
	)
}

type wsaPage struct {
	// 标题
	Title string `json:"title,omitempty"`
	// 简介
	Passage string `json:"passage,omitempty"`
	// 发布日期
	Date string `json:"date,omitempty"`
	// 网站 URL
	URL string `json:"url,omitempty"`
	// 所属网站
	Site string `json:"site,omitempty"`
	// 相关性评分
	Score float64 `json:"score,omitempty"`
}
