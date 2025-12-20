package agents

import (
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/shopspring/decimal"
)

// QueryAssetPriceTrendsRequest 查询资产价格趋势请求
type QueryAssetPriceTrendsRequest struct {
	// 资产代号
	Code string `json:"code"`
}

// QueryAssetPriceTrendsResponse 查询资产价格趋势响应
type QueryAssetPriceTrendsResponse struct {
	Data []TimeSeriesItem `json:"data"`
}

// TimeSeriesItem 时序数据项
type TimeSeriesItem struct {
	Timestamp int64  `json:"ts"`
	Value     string `json:"value"`
}

const QueryAssetPriceTrendsToolName = "QueryAssetPriceTrends"

// QueryAssetPriceTrends 查询资产价格
//
// TODO: 这是一个假实现，仅做调试用
func QueryAssetPriceTrends(
	_ *ai.ToolContext,
	_ QueryAssetPriceTrendsRequest,
) (QueryAssetPriceTrendsResponse, error) {
	now := time.Now()
	return QueryAssetPriceTrendsResponse{
		Data: []TimeSeriesItem{
			{Timestamp: now.Add(6 * 24 * time.Hour).Unix(), Value: decimal.New(52200, -2).String()},
			{Timestamp: now.Add(5 * 24 * time.Hour).Unix(), Value: decimal.New(51300, -2).String()},
			{Timestamp: now.Add(4 * 24 * time.Hour).Unix(), Value: decimal.New(55000, -2).String()},
			{Timestamp: now.Add(3 * 24 * time.Hour).Unix(), Value: decimal.New(51000, -2).String()},
			{Timestamp: now.Add(2 * 24 * time.Hour).Unix(), Value: decimal.New(52000, -2).String()},
			{Timestamp: now.Add(24 * time.Hour).Unix(), Value: decimal.New(51000, -2).String()},
			{Timestamp: now.Unix(), Value: decimal.New(50100, -2).String()},
		},
	}, nil
}

// DefineToolQueryAssetPriceTrends 定义查询资产价格工具
func DefineToolQueryAssetPriceTrends(g *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(
		g, QueryAssetPriceTrendsToolName,
		"Query the price trends of assets such as stocks, funds, bonds, ETFs, etc.",
		QueryAssetPriceTrends,
	)
}
