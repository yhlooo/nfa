package trading

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosition_TotalValue(t *testing.T) {
	tests := []struct {
		name      string
		position  *Position
		yesBid    float64
		noBid     float64
		wantValue float64
	}{
		{
			name:      "empty position",
			position:  NewPosition(),
			yesBid:    0.5,
			noBid:     0.5,
			wantValue: 0,
		},
		{
			name: "only cash",
			position: &Position{
				Cash: 100,
			},
			yesBid:    0.5,
			noBid:     0.5,
			wantValue: 100,
		},
		{
			name: "only yes shares",
			position: &Position{
				YesShares: 100,
			},
			yesBid:    0.5,
			noBid:     0.5,
			wantValue: 50,
		},
		{
			name: "only no shares",
			position: &Position{
				NoShares: 100,
			},
			yesBid:    0.5,
			noBid:     0.6,
			wantValue: 60,
		},
		{
			name: "mixed position - positive",
			position: &Position{
				Cash:      -50,
				YesShares: 100,
				NoShares:  50,
			},
			yesBid:    0.6,
			noBid:     0.4,
			wantValue: -50 + 100*0.6 + 50*0.4, // -50 + 60 + 20 = 30
		},
		{
			name: "mixed position - negative",
			position: &Position{
				Cash:      -100,
				YesShares: 50,
				NoShares:  0,
			},
			yesBid:    0.3,
			noBid:     0.7,
			wantValue: -100 + 50*0.3, // -100 + 15 = -85
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.position.TotalValue(tt.yesBid, tt.noBid)
			assert.Equal(t, tt.wantValue, got)
		})
	}
}

func TestPosition_BuySell(t *testing.T) {
	p := NewPosition()

	// 买入 Yes
	p.BuyYes(100, 0.4)
	assert.Equal(t, -40.0, p.Cash)
	assert.Equal(t, 100.0, p.YesShares)
	assert.Equal(t, 0.4, p.YesAvgCost)

	// 再买入 Yes（不同价格）
	p.BuyYes(100, 0.5)
	assert.Equal(t, -90.0, p.Cash)
	assert.Equal(t, 200.0, p.YesShares)
	assert.Equal(t, 0.45, p.YesAvgCost) // (40+50)/200

	// 卖出部分 Yes
	p.SellYes(50, 0.6)
	assert.Equal(t, -60.0, p.Cash) // -90 + 30
	assert.Equal(t, 150.0, p.YesShares)
	assert.Equal(t, 0.45, p.YesAvgCost) // 平均成本不变

	// 卖出所有 Yes
	p.SellYes(200, 0.7)                     // 超过持仓，应该只卖 150
	assert.InDelta(t, 45.0, p.Cash, 0.0001) // -60 + 150*0.7
	assert.Equal(t, 0.0, p.YesShares)
	assert.Equal(t, 0.0, p.YesAvgCost) // 清空后重置
}

func TestPosition_Clone(t *testing.T) {
	original := &Position{
		Cash:       100,
		YesShares:  50,
		NoShares:   30,
		YesAvgCost: 0.4,
		NoAvgCost:  0.6,
	}

	clone := original.Clone()

	// 验证值相同
	assert.Equal(t, original, clone)

	// 修改 clone 不影响 original
	clone.Cash = 200
	assert.Equal(t, 100.0, original.Cash)
}
