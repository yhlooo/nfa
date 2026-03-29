package trading

import (
	"time"
)

// PricePoint 价格点
type PricePoint struct {
	Time  time.Time
	Value float64
}

// PriceHistory 价格历史管理
type PriceHistory struct {
	maxSize int

	yesBids []PricePoint
	yesAsks []PricePoint
	noBids  []PricePoint
	noAsks  []PricePoint
}

// NewPriceHistory 创建价格历史
func NewPriceHistory(maxSize int) *PriceHistory {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &PriceHistory{
		maxSize: maxSize,
	}
}

// AddYesBid 添加 Yes 买价
func (h *PriceHistory) AddYesBid(point PricePoint) {
	h.yesBids = append(h.yesBids, point)
	if len(h.yesBids) > h.maxSize {
		h.yesBids = h.yesBids[len(h.yesBids)-h.maxSize:]
	}
}

// AddYesAsk 添加 Yes 卖价
func (h *PriceHistory) AddYesAsk(point PricePoint) {
	h.yesAsks = append(h.yesAsks, point)
	if len(h.yesAsks) > h.maxSize {
		h.yesAsks = h.yesAsks[len(h.yesAsks)-h.maxSize:]
	}
}

// AddNoBid 添加 No 买价
func (h *PriceHistory) AddNoBid(point PricePoint) {
	h.noBids = append(h.noBids, point)
	if len(h.noBids) > h.maxSize {
		h.noBids = h.noBids[len(h.noBids)-h.maxSize:]
	}
}

// AddNoAsk 添加 No 卖价
func (h *PriceHistory) AddNoAsk(point PricePoint) {
	h.noAsks = append(h.noAsks, point)
	if len(h.noAsks) > h.maxSize {
		h.noAsks = h.noAsks[len(h.noAsks)-h.maxSize:]
	}
}

// YesBids 获取 Yes 买价历史
func (h *PriceHistory) YesBids() []PricePoint {
	return h.yesBids
}

// YesAsks 获取 Yes 卖价历史
func (h *PriceHistory) YesAsks() []PricePoint {
	return h.yesAsks
}

// NoBids 获取 No 买价历史
func (h *PriceHistory) NoBids() []PricePoint {
	return h.noBids
}

// NoAsks 获取 No 卖价历史
func (h *PriceHistory) NoAsks() []PricePoint {
	return h.noAsks
}

// LatestYesBid 获取最新 Yes 买价
func (h *PriceHistory) LatestYesBid() *float64 {
	if len(h.yesBids) == 0 {
		return nil
	}
	v := h.yesBids[len(h.yesBids)-1].Value
	return &v
}

// LatestYesAsk 获取最新 Yes 卖价
func (h *PriceHistory) LatestYesAsk() *float64 {
	if len(h.yesAsks) == 0 {
		return nil
	}
	v := h.yesAsks[len(h.yesAsks)-1].Value
	return &v
}

// LatestNoBid 获取最新 No 买价
func (h *PriceHistory) LatestNoBid() *float64 {
	if len(h.noBids) == 0 {
		return nil
	}
	v := h.noBids[len(h.noBids)-1].Value
	return &v
}

// LatestNoAsk 获取最新 No 卖价
func (h *PriceHistory) LatestNoAsk() *float64 {
	if len(h.noAsks) == 0 {
		return nil
	}
	v := h.noAsks[len(h.noAsks)-1].Value
	return &v
}
