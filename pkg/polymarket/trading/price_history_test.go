package trading

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPriceHistory_AddAndGet(t *testing.T) {
	h := NewPriceHistory(5)

	now := time.Now()

	// 添加 Yes bid 价格
	for i := 0; i < 3; i++ {
		h.AddYesBid(PricePoint{Time: now.Add(time.Duration(i) * time.Second), Value: float64(i) + 0.1})
	}

	// 添加 Yes ask 价格
	for i := 0; i < 3; i++ {
		h.AddYesAsk(PricePoint{Time: now.Add(time.Duration(i) * time.Second), Value: float64(i) + 0.2})
	}

	// 验证数量
	assert.Len(t, h.YesBids(), 3)
	assert.Len(t, h.YesAsks(), 3)
	assert.Len(t, h.NoBids(), 0)
	assert.Len(t, h.NoAsks(), 0)

	// 验证最新价格
	latestBid := h.LatestYesBid()
	assert.NotNil(t, latestBid)
	assert.Equal(t, 2.1, *latestBid)

	latestAsk := h.LatestYesAsk()
	assert.NotNil(t, latestAsk)
	assert.Equal(t, 2.2, *latestAsk)

	// 空 No 价格
	assert.Nil(t, h.LatestNoBid())
	assert.Nil(t, h.LatestNoAsk())
}

func TestPriceHistory_MaxSize(t *testing.T) {
	maxSize := 5
	h := NewPriceHistory(maxSize)

	now := time.Now()

	// 添加超过 maxSize 的价格
	for i := 0; i < 10; i++ {
		h.AddYesBid(PricePoint{Time: now.Add(time.Duration(i) * time.Second), Value: float64(i)})
	}

	// 应该只保留最后 maxSize 个
	assert.Len(t, h.YesBids(), maxSize)

	// 验证保留的是最新的
	bids := h.YesBids()
	assert.Equal(t, 5.0, bids[0].Value) // 第 6 个 (index 5)
	assert.Equal(t, 9.0, bids[4].Value) // 第 10 个 (index 9)

	// 最新价格是最后一个
	latest := h.LatestYesBid()
	assert.NotNil(t, latest)
	assert.Equal(t, 9.0, *latest)
}

func TestPriceHistory_DefaultMaxSize(t *testing.T) {
	// 传入 0 或负数应该使用默认值 1000
	h1 := NewPriceHistory(0)
	assert.Equal(t, 1000, h1.maxSize)

	h2 := NewPriceHistory(-1)
	assert.Equal(t, 1000, h2.maxSize)

	h3 := NewPriceHistory(100)
	assert.Equal(t, 100, h3.maxSize)
}

func TestPriceHistory_AllPriceTypes(t *testing.T) {
	h := NewPriceHistory(10)
	now := time.Now()

	// 添加所有类型
	h.AddYesBid(PricePoint{Time: now, Value: 0.1})
	h.AddYesAsk(PricePoint{Time: now, Value: 0.2})
	h.AddNoBid(PricePoint{Time: now, Value: 0.3})
	h.AddNoAsk(PricePoint{Time: now, Value: 0.4})

	// 验证所有类型都有数据
	assert.Len(t, h.YesBids(), 1)
	assert.Len(t, h.YesAsks(), 1)
	assert.Len(t, h.NoBids(), 1)
	assert.Len(t, h.NoAsks(), 1)

	// 验证最新价格
	assert.Equal(t, 0.1, *h.LatestYesBid())
	assert.Equal(t, 0.2, *h.LatestYesAsk())
	assert.Equal(t, 0.3, *h.LatestNoBid())
	assert.Equal(t, 0.4, *h.LatestNoAsk())
}
