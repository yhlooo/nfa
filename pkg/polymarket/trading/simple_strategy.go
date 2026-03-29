package trading

import (
	"context"
	"math"
)

// 最小买入价值（美元）
const minBuyValue = 1.0

// SimpleStrategy 简单示例策略
//
// 规则：
// - Yes 或 No 价格低于 0.3 时买入，最多持有一单
// - 持仓价格高于 0.5 时卖出
// - 卖出后可以再买
// - 买入数量确保价值至少 1 美元
type SimpleStrategy struct{}

// NewSimpleStrategy 创建简单策略
func NewSimpleStrategy() *SimpleStrategy {
	return &SimpleStrategy{}
}

// Name 返回策略名称
func (s *SimpleStrategy) Name() string {
	return "simple"
}

// Execute 执行策略
func (s *SimpleStrategy) Execute(ctx context.Context, in Input) (*Result, error) {
	result := NewResult()

	// 获取最新价格
	yesAsk := latestValue(in.YesAskPrices)
	yesBid := latestValue(in.YesBidPrices)
	noAsk := latestValue(in.NoAskPrices)
	noBid := latestValue(in.NoBidPrices)

	if yesAsk == nil || yesBid == nil || noAsk == nil || noBid == nil {
		return result, nil
	}

	// 检查是否有持仓（最多持有一单）
	hasYesPosition := in.Position.YesShares > 0
	hasNoPosition := in.Position.NoShares > 0

	// 规则 1: 卖出 - 价格高于 0.5 时卖出
	if hasYesPosition && *yesBid > 0.5 {
		result.AddOrder(NewOrderRequest(
			OrderSideSell,
			OutcomeYes,
			in.Position.YesShares,
			*yesBid,
			OrderTypeMarket,
		))
		return result, nil
	}

	if hasNoPosition && *noBid > 0.5 {
		result.AddOrder(NewOrderRequest(
			OrderSideSell,
			OutcomeNo,
			in.Position.NoShares,
			*noBid,
			OrderTypeMarket,
		))
		return result, nil
	}

	// 规则 2: 买入 - 价格低于 0.3 且没有持仓时买入
	if !hasYesPosition && !hasNoPosition {
		// 优先买入价格更低的
		if *yesAsk < 0.3 && *yesAsk <= *noAsk {
			size := calcMinBuySize(*yesAsk)
			result.AddOrder(NewOrderRequest(
				OrderSideBuy,
				OutcomeYes,
				size,
				*yesAsk,
				OrderTypeMarket,
			))
			return result, nil
		}

		if *noAsk < 0.3 {
			size := calcMinBuySize(*noAsk)
			result.AddOrder(NewOrderRequest(
				OrderSideBuy,
				OutcomeNo,
				size,
				*noAsk,
				OrderTypeMarket,
			))
			return result, nil
		}
	}

	return result, nil
}

// calcMinBuySize 计算最小买入份额，确保价值至少 minBuyValue 美元
func calcMinBuySize(price float64) float64 {
	if price <= 0 {
		return 0
	}
	// size * price >= minBuyValue
	// size >= minBuyValue / price
	return math.Ceil(minBuyValue/price*100) / 100 // 向上取整到 0.01
}

// latestValue 获取最新价格
func latestValue(points []PricePoint) *float64 {
	if len(points) == 0 {
		return nil
	}
	v := points[len(points)-1].Value
	return &v
}
