package trading

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleStrategy_Name(t *testing.T) {
	s := NewSimpleStrategy()
	assert.Equal(t, "simple", s.Name())
}

func TestSimpleStrategy_BuyWhenPriceLow(t *testing.T) {
	s := NewSimpleStrategy()
	ctx := context.Background()

	tests := []struct {
		name          string
		input         Input
		expectOrder   bool
		expectSide    OrderSide
		expectOutcome Outcome
		expectPrice   float64
	}{
		{
			name: "buy yes when yes ask < 0.3 and cheaper than no",
			input: Input{
				Position:     NewPosition(),
				YesAskPrices: []PricePoint{{Value: 0.25}},
				YesBidPrices: []PricePoint{{Value: 0.24}},
				NoAskPrices:  []PricePoint{{Value: 0.30}},
				NoBidPrices:  []PricePoint{{Value: 0.29}},
			},
			expectOrder:   true,
			expectSide:    OrderSideBuy,
			expectOutcome: OutcomeYes,
			expectPrice:   0.25,
		},
		{
			name: "buy no when no ask < 0.3 and cheaper than yes",
			input: Input{
				Position:     NewPosition(),
				YesAskPrices: []PricePoint{{Value: 0.30}},
				YesBidPrices: []PricePoint{{Value: 0.29}},
				NoAskPrices:  []PricePoint{{Value: 0.25}},
				NoBidPrices:  []PricePoint{{Value: 0.24}},
			},
			expectOrder:   true,
			expectSide:    OrderSideBuy,
			expectOutcome: OutcomeNo,
			expectPrice:   0.25,
		},
		{
			name: "buy yes when both < 0.3 but yes is cheaper",
			input: Input{
				Position:     NewPosition(),
				YesAskPrices: []PricePoint{{Value: 0.20}},
				YesBidPrices: []PricePoint{{Value: 0.19}},
				NoAskPrices:  []PricePoint{{Value: 0.25}},
				NoBidPrices:  []PricePoint{{Value: 0.24}},
			},
			expectOrder:   true,
			expectSide:    OrderSideBuy,
			expectOutcome: OutcomeYes,
			expectPrice:   0.20,
		},
		{
			name: "no buy when price >= 0.3",
			input: Input{
				Position:     NewPosition(),
				YesAskPrices: []PricePoint{{Value: 0.35}},
				YesBidPrices: []PricePoint{{Value: 0.34}},
				NoAskPrices:  []PricePoint{{Value: 0.40}},
				NoBidPrices:  []PricePoint{{Value: 0.39}},
			},
			expectOrder: false,
		},
		{
			name: "no buy when already has yes position",
			input: Input{
				Position: &Position{
					YesShares:  10,
					YesAvgCost: 0.25,
				},
				YesAskPrices: []PricePoint{{Value: 0.20}},
				YesBidPrices: []PricePoint{{Value: 0.19}},
				NoAskPrices:  []PricePoint{{Value: 0.25}},
				NoBidPrices:  []PricePoint{{Value: 0.24}},
			},
			expectOrder: false,
		},
		{
			name: "no buy when already has no position",
			input: Input{
				Position: &Position{
					NoShares:  10,
					NoAvgCost: 0.25,
				},
				YesAskPrices: []PricePoint{{Value: 0.20}},
				YesBidPrices: []PricePoint{{Value: 0.19}},
				NoAskPrices:  []PricePoint{{Value: 0.25}},
				NoBidPrices:  []PricePoint{{Value: 0.24}},
			},
			expectOrder: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.Execute(ctx, tt.input)
			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.expectOrder {
				require.Len(t, result.Orders, 1)
				order := result.Orders[0]
				assert.Equal(t, tt.expectSide, order.Side)
				assert.Equal(t, tt.expectOutcome, order.Outcome)
				assert.Equal(t, 1.0, order.Size)
				assert.Equal(t, tt.expectPrice, order.Price)
				assert.Equal(t, OrderTypeMarket, order.OrderType)
			} else {
				assert.True(t, result.IsEmpty())
			}
		})
	}
}

func TestSimpleStrategy_SellWhenPriceHigh(t *testing.T) {
	s := NewSimpleStrategy()
	ctx := context.Background()

	tests := []struct {
		name          string
		input         Input
		expectOrder   bool
		expectSide    OrderSide
		expectOutcome Outcome
		expectSize    float64
	}{
		{
			name: "sell yes when bid > 0.5",
			input: Input{
				Position: &Position{
					YesShares:  10,
					YesAvgCost: 0.25,
				},
				YesAskPrices: []PricePoint{{Value: 0.55}},
				YesBidPrices: []PricePoint{{Value: 0.52}},
				NoAskPrices:  []PricePoint{{Value: 0.50}},
				NoBidPrices:  []PricePoint{{Value: 0.48}},
			},
			expectOrder:   true,
			expectSide:    OrderSideSell,
			expectOutcome: OutcomeYes,
			expectSize:    10,
		},
		{
			name: "sell no when bid > 0.5",
			input: Input{
				Position: &Position{
					NoShares:  15,
					NoAvgCost: 0.30,
				},
				YesAskPrices: []PricePoint{{Value: 0.50}},
				YesBidPrices: []PricePoint{{Value: 0.48}},
				NoAskPrices:  []PricePoint{{Value: 0.55}},
				NoBidPrices:  []PricePoint{{Value: 0.52}},
			},
			expectOrder:   true,
			expectSide:    OrderSideSell,
			expectOutcome: OutcomeNo,
			expectSize:    15,
		},
		{
			name: "no sell when bid <= 0.5",
			input: Input{
				Position: &Position{
					YesShares:  10,
					YesAvgCost: 0.25,
				},
				YesAskPrices: []PricePoint{{Value: 0.45}},
				YesBidPrices: []PricePoint{{Value: 0.42}},
				NoAskPrices:  []PricePoint{{Value: 0.60}},
				NoBidPrices:  []PricePoint{{Value: 0.58}},
			},
			expectOrder: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.Execute(ctx, tt.input)
			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.expectOrder {
				require.Len(t, result.Orders, 1)
				order := result.Orders[0]
				assert.Equal(t, tt.expectSide, order.Side)
				assert.Equal(t, tt.expectOutcome, order.Outcome)
				assert.Equal(t, tt.expectSize, order.Size)
				assert.Equal(t, OrderTypeMarket, order.OrderType)
			} else {
				assert.True(t, result.IsEmpty())
			}
		})
	}
}

func TestSimpleStrategy_SellThenBuy(t *testing.T) {
	s := NewSimpleStrategy()
	ctx := context.Background()

	// 场景：持有 Yes，价格上涨到 0.5 以上卖出
	input1 := Input{
		Position: &Position{
			YesShares:  10,
			YesAvgCost: 0.25,
		},
		YesAskPrices: []PricePoint{{Value: 0.55}},
		YesBidPrices: []PricePoint{{Value: 0.52}},
		NoAskPrices:  []PricePoint{{Value: 0.50}},
		NoBidPrices:  []PricePoint{{Value: 0.48}},
	}

	result1, err := s.Execute(ctx, input1)
	require.NoError(t, err)
	require.Len(t, result1.Orders, 1)
	assert.Equal(t, OrderSideSell, result1.Orders[0].Side)
	assert.Equal(t, OutcomeYes, result1.Orders[0].Outcome)

	// 卖出后，可以再买入（模拟卖出后的状态）
	input2 := Input{
		Position:     NewPosition(), // 空仓
		YesAskPrices: []PricePoint{{Value: 0.25}},
		YesBidPrices: []PricePoint{{Value: 0.24}},
		NoAskPrices:  []PricePoint{{Value: 0.50}},
		NoBidPrices:  []PricePoint{{Value: 0.48}},
	}

	result2, err := s.Execute(ctx, input2)
	require.NoError(t, err)
	require.Len(t, result2.Orders, 1)
	assert.Equal(t, OrderSideBuy, result2.Orders[0].Side)
}

func TestSimpleStrategy_NoPriceData(t *testing.T) {
	s := NewSimpleStrategy()
	ctx := context.Background()

	input := Input{
		Position:     NewPosition(),
		YesAskPrices: nil,
		YesBidPrices: nil,
		NoAskPrices:  nil,
		NoBidPrices:  nil,
	}

	result, err := s.Execute(ctx, input)
	require.NoError(t, err)
	assert.True(t, result.IsEmpty())
}
