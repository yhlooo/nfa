package trading

import (
	"context"
	"math/rand"
	"time"
)

// RandomStrategy 随机策略
//
// 规则：
// - 每 5 秒执行一次随机操作
// - 随机选择买入或卖出
// - 随机选择 Yes 或 No
// - 买入时数量确保价值至少 1 美元
// - 卖出时卖出全部持仓（如果有）
type RandomStrategy struct {
	rng     *rand.Rand
	lastAct time.Time
}

// NewRandomStrategy 创建随机策略
func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Name 返回策略名称
func (s *RandomStrategy) Name() string {
	return "rand"
}

// Execute 执行策略
func (s *RandomStrategy) Execute(ctx context.Context, in Input) (*Result, error) {
	result := NewResult()

	// 获取最新价格
	yesAsk := latestValue(in.YesAskPrices)
	yesBid := latestValue(in.YesBidPrices)
	noAsk := latestValue(in.NoAskPrices)
	noBid := latestValue(in.NoBidPrices)

	if yesAsk == nil || yesBid == nil || noAsk == nil || noBid == nil {
		return result, nil
	}

	// 每 5 秒执行一次
	now := time.Now()
	if now.Sub(s.lastAct) < 5*time.Second {
		return result, nil
	}
	s.lastAct = now

	// 随机选择操作：0 = 买入, 1 = 卖出
	action := s.rng.Intn(2)

	// 随机选择结果：0 = Yes, 1 = No
	outcome := s.rng.Intn(2)

	if action == 0 {
		// 买入
		if outcome == 0 {
			size := calcMinBuySize(*yesAsk)
			result.AddOrder(NewOrderRequest(
				OrderSideBuy,
				OutcomeYes,
				size,
				*yesAsk,
				OrderTypeMarket,
			))
		} else {
			size := calcMinBuySize(*noAsk)
			result.AddOrder(NewOrderRequest(
				OrderSideBuy,
				OutcomeNo,
				size,
				*noAsk,
				OrderTypeMarket,
			))
		}
	} else {
		// 卖出（需要有持仓）
		if outcome == 0 && in.Position.YesShares > 0 {
			result.AddOrder(NewOrderRequest(
				OrderSideSell,
				OutcomeYes,
				in.Position.YesShares,
				*yesBid,
				OrderTypeMarket,
			))
		} else if outcome == 1 && in.Position.NoShares > 0 {
			result.AddOrder(NewOrderRequest(
				OrderSideSell,
				OutcomeNo,
				in.Position.NoShares,
				*noBid,
				OrderTypeMarket,
			))
		}
		// 如果没有对应持仓，本次不操作
	}

	return result, nil
}
