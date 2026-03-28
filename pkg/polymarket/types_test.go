package polymarket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResolutionSource(t *testing.T) {
	t.Run("Chainlink BTC-USD", func(t *testing.T) {
		market := &Market{
			Events: []MarketEventData{
				{ResolutionSource: "https://data.chain.link/streams/btc-usd"},
			},
		}
		info := ParseResolutionSource(market)
		assert.NotNil(t, info)
		assert.Equal(t, "btc", info.Asset)
		assert.Equal(t, "usd", info.Quote)
		assert.Equal(t, "crypto_prices_chainlink", info.Topic)
		assert.Equal(t, "btc/usd", info.Symbol)
	})

	t.Run("Chainlink ETH-USD", func(t *testing.T) {
		market := &Market{
			Events: []MarketEventData{
				{ResolutionSource: "https://data.chain.link/streams/eth-usd"},
			},
		}
		info := ParseResolutionSource(market)
		assert.NotNil(t, info)
		assert.Equal(t, "eth", info.Asset)
		assert.Equal(t, "usd", info.Quote)
		assert.Equal(t, "eth/usd", info.Symbol)
	})

	t.Run("empty resolutionSource", func(t *testing.T) {
		market := &Market{
			Events: []MarketEventData{
				{ResolutionSource: ""},
			},
		}
		info := ParseResolutionSource(market)
		assert.Nil(t, info)
	})

	t.Run("unmatched URL", func(t *testing.T) {
		market := &Market{
			Events: []MarketEventData{
				{ResolutionSource: "https://example.com/other-source"},
			},
		}
		info := ParseResolutionSource(market)
		assert.Nil(t, info)
	})

	t.Run("nil market", func(t *testing.T) {
		info := ParseResolutionSource(nil)
		assert.Nil(t, info)
	})

	t.Run("empty events", func(t *testing.T) {
		market := &Market{}
		info := ParseResolutionSource(market)
		assert.Nil(t, info)
	})
}
