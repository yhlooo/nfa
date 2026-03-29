package trading

// MarketInfo 市场元信息
type MarketInfo struct {
	// ID 市场 ID
	ID string
	// Slug 市场 slug
	Slug string
	// Question 市场问题
	Question string
	// Description 市场描述
	Description string
	// YesAssetID Yes 资产 ID
	YesAssetID string
	// NoAssetID No 资产 ID
	NoAssetID string
}
