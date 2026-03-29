package trading

// Position 持仓信息
type Position struct {
	// Cash 现金（可负）
	Cash float64
	// YesShares Yes 持仓数量
	YesShares float64
	// NoShares No 持仓数量
	NoShares float64
	// YesAvgCost Yes 平均成本
	YesAvgCost float64
	// NoAvgCost No 平均成本
	NoAvgCost float64
}

// NewPosition 创建空持仓
func NewPosition() *Position {
	return &Position{}
}

// TotalValue 计算资产总价值
// 总价值 = 现金 + Yes持仓 * Yes买价 + No持仓 * No买价
func (p *Position) TotalValue(yesBid, noBid float64) float64 {
	return p.Cash + p.YesShares*yesBid + p.NoShares*noBid
}

// Clone 克隆持仓
func (p *Position) Clone() *Position {
	return &Position{
		Cash:       p.Cash,
		YesShares:  p.YesShares,
		NoShares:   p.NoShares,
		YesAvgCost: p.YesAvgCost,
		NoAvgCost:  p.NoAvgCost,
	}
}

// BuyYes 买入 Yes
func (p *Position) BuyYes(size, price float64) {
	cost := size * price
	p.Cash -= cost

	// 更新平均成本
	totalCost := p.YesAvgCost*p.YesShares + cost
	p.YesShares += size
	if p.YesShares > 0 {
		p.YesAvgCost = totalCost / p.YesShares
	}
}

// SellYes 卖出 Yes
func (p *Position) SellYes(size, price float64) {
	if size > p.YesShares {
		size = p.YesShares
	}
	p.Cash += size * price
	p.YesShares -= size

	// 清空时重置平均成本
	if p.YesShares == 0 {
		p.YesAvgCost = 0
	}
}

// BuyNo 买入 No
func (p *Position) BuyNo(size, price float64) {
	cost := size * price
	p.Cash -= cost

	// 更新平均成本
	totalCost := p.NoAvgCost*p.NoShares + cost
	p.NoShares += size
	if p.NoShares > 0 {
		p.NoAvgCost = totalCost / p.NoShares
	}
}

// SellNo 卖出 No
func (p *Position) SellNo(size, price float64) {
	if size > p.NoShares {
		size = p.NoShares
	}
	p.Cash += size * price
	p.NoShares -= size

	// 清空时重置平均成本
	if p.NoShares == 0 {
		p.NoAvgCost = 0
	}
}
