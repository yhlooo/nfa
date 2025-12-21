package agents

import (
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/shopspring/decimal"
)

// CheckAssetPriceTrendsRequest 查询资产价格趋势请求
type CheckAssetPriceTrendsRequest struct {
	// 资产代号
	Code string `json:"code"`
}

// CheckAssetPriceTrendsResponse 查询资产价格趋势响应
type CheckAssetPriceTrendsResponse struct {
	Data []TimeSeriesItem `json:"data"`
}

// TimeSeriesItem 时序数据项
type TimeSeriesItem struct {
	Timestamp int64  `json:"ts"`
	Value     string `json:"value"`
}

const CheckAssetPriceTrendsToolName = "CheckAssetPriceTrends"

// CheckAssetPriceTrends 查询资产价格
//
// TODO: 这是一个假实现，仅做调试用
func CheckAssetPriceTrends(
	_ *ai.ToolContext,
	_ CheckAssetPriceTrendsRequest,
) (CheckAssetPriceTrendsResponse, error) {
	now := time.Now()
	return CheckAssetPriceTrendsResponse{
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

// DefineToolCheckAssetPriceTrends 定义查询资产价格工具
func DefineToolCheckAssetPriceTrends(g *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(
		g, CheckAssetPriceTrendsToolName,
		"Check the price trends of assets such as stocks, funds, bonds, ETFs, etc.",
		CheckAssetPriceTrends,
	)
}
