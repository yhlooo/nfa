package dataproviders

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/mcp"
)

const (
	AlphaVantageMCPBaseURL = "https://mcp.alphavantage.co/mcp"
)

// AlphaVantageOptions AlphaVantage 选项
type AlphaVantageOptions struct {
	APIKey string `json:"apiKey"`
}

// RegisterTools 注册工具
func (opts *AlphaVantageOptions) RegisterTools(ctx context.Context, g *genkit.Genkit) (
	comprehensiveAnalysisTools []ai.ToolRef,
	macroeconomicAnalysisTools []ai.ToolRef,
	fundamentalAnalysisTools []ai.ToolRef,
	technicalAnalysisTools []ai.ToolRef,
	allTools []ai.ToolRef,
	err error,
) {
	client, err := mcp.NewGenkitMCPClient(mcp.MCPClientOptions{
		Name: "alpha-vantage",
		StreamableHTTP: &mcp.StreamableHTTPConfig{
			BaseURL: AlphaVantageMCPBaseURL + "?apikey=" + opts.APIKey,
		},
	})
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("init alpha vantage mcp client error: %w", err)
	}

	tools, err := client.GetActiveTools(ctx, g)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("get active tools error: %w", err)
	}

	for _, tool := range tools {
		desc := tool.Definition()
		var toolOpts []ai.ToolOption
		if len(desc.InputSchema) > 0 {
			toolOpts = append(toolOpts, ai.WithInputSchema(desc.InputSchema))
		}
		genkit.DefineTool(g, desc.Name, desc.Description, MCPToolFn(tool.RunRaw), toolOpts...)

		allTools = append(allTools, tool)

		switch desc.Name {
		case "alpha-vantage_WTI", // WTI 原油价格
			"alpha-vantage_BRENT",           // 布伦特原油价格
			"alpha-vantage_NATURAL_GAS",     // 亨利中心天然气现货价格
			"alpha-vantage_COPPER",          // 全球铜价
			"alpha-vantage_ALUMINUM",        // 全球铝价
			"alpha-vantage_WHEAT",           // 全球小麦价格
			"alpha-vantage_CORN",            // 全球玉米价格
			"alpha-vantage_COTTON",          // 全球棉花价格
			"alpha-vantage_SUGAR",           // 全球糖价
			"alpha-vantage_COFFEE",          // 全球咖啡价格
			"alpha-vantage_ALL_COMMODITIES", // 所有商品价格

			"alpha-vantage_REAL_GDP",            // 实际国内生产总值
			"alpha-vantage_REAL_GDP_PER_CAPITA", // 人均实际 GDP
			"alpha-vantage_TREASURY_YIELD",      // 日国债收益率率
			"alpha-vantage_FEDERAL_FUNDS_RATE",  // 联邦基金利率（利率）
			"alpha-vantage_CPI",                 // 消费者物价指数
			"alpha-vantage_INFLATION",           // 通货膨胀率
			"alpha-vantage_RETAIL_SALES",        // 零售销售数据
			"alpha-vantage_DURABLES",            // 耐用品订单
			"alpha-vantage_UNEMPLOYMENT",        // 失业率
			"alpha-vantage_NONFARM_PAYROLL":     // 非农就业数据
			macroeconomicAnalysisTools = append(macroeconomicAnalysisTools, tool)

		case "alpha-vantage_SMA", // 简单移动平均线（SMA）值
			"alpha-vantage_EMA",          // 指数移动平均线（EMA）值
			"alpha-vantage_WMA",          // 加权移动平均线（WMA）数值
			"alpha-vantage_DEMA",         // 双指数移动平均（DEMA）数值
			"alpha-vantage_TEMA",         // 三重指数移动平均（TEMA）值
			"alpha-vantage_TRIMA",        // 三角移动平均（TRIMA）值
			"alpha-vantage_KAMA",         // 考夫曼自适应移动平均（KAMA）值
			"alpha-vantage_MAMA",         // MESA自适应移动平均（MAMA）值
			"alpha-vantage_VWAP",         // 盘中时间序列的成交量加权平均价格（VWAP）
			"alpha-vantage_T3",           // 三重指数移动平均线（T3）值
			"alpha-vantage_MACD",         // 移动平均收敛/背离（MACD）值
			"alpha-vantage_MACDEXT",      // 具有可控移动平均类型的移动平均收敛/背离值
			"alpha-vantage_STOCH",        // 随机振荡器（STOCH）值
			"alpha-vantage_STOCHF",       // 随机快速（STOCHF）值
			"alpha-vantage_RSI",          // 相对强弱指数（RSI）值
			"alpha-vantage_STOCHRSI",     // 随机相对强度指数（STOCHRSI）值
			"alpha-vantage_WILLR",        // 威廉姆斯的%R（WILLR）值
			"alpha-vantage_ADX",          // 平均方向移动指数（ADX）值
			"alpha-vantage_ADXR",         // 平均方向移动指数评级（ADXR）值
			"alpha-vantage_APO",          // 绝对价格振荡器（APO）值
			"alpha-vantage_PPO",          // 百分比价格振荡器（PPO）值
			"alpha-vantage_MOM",          // 动量（MOM）值
			"alpha-vantage_BOP",          // 权力平衡（BOP）值
			"alpha-vantage_CCI",          // 商品通道指数（CCI）价值
			"alpha-vantage_CMO",          // 尚德动量振荡器（CMO）值
			"alpha-vantage_ROC",          // 变化率（ROC）值
			"alpha-vantage_ROCR",         // 变化率（ROCR）值
			"alpha-vantage_AROON",        // Aroon（AROON）价值观
			"alpha-vantage_AROONOSC",     // Aroon 振荡器（AROONOSC）值
			"alpha-vantage_MFI",          // 资金流动指数（MFI）价值
			"alpha-vantage_TRIX",         // 三重平滑指数移动平均线（TRIX）值的1天变化率
			"alpha-vantage_ULTOSC",       // 极极振荡器（ULTOSC）值
			"alpha-vantage_DX",           // 方向移动指数（DX）值
			"alpha-vantage_MINUS_DI",     // 负方向指示（MINUS_DI）值
			"alpha-vantage_PLUS_DI",      // 加上方向指示器（PLUS_DI）值
			"alpha-vantage_MINUS_DM",     // 负方向运动（MINUS_DM）值
			"alpha-vantage_PLUS_DM",      // 加上方向移动（PLUS_DM）值
			"alpha-vantage_BBANDS",       // 布林带（BBANDS）值
			"alpha-vantage_MIDPOINT",     // 中点值 - （最高值 + 最低值）/2
			"alpha-vantage_MIDPRICE",     // 中点价格 - （最高价 + 最低价）/2
			"alpha-vantage_SAR",          // 抛物线SAR（SAR）值
			"alpha-vantage_TRANGE",       // 真实距离（TRANGE）值
			"alpha-vantage_ATR",          // 平均真实射程（ATR）值
			"alpha-vantage_NATR",         // 归一化平均真实距离（NATR）值
			"alpha-vantage_AD",           // 柴金A/D线（AD）值
			"alpha-vantage_ADOSC",        // Chaikin A/D 振荡器（ADOSC）值
			"alpha-vantage_OBV",          // 按平衡量（OBV）值
			"alpha-vantage_HT_TRENDLINE", // 希尔伯特变换，瞬时趋势线（HT_TRENDLINE）值
			"alpha-vantage_HT_SINE",      // 希尔伯特变换，正弦波（HT_SINE）值
			"alpha-vantage_HT_TRENDMODE", // 希尔伯特变换，趋势与循环模式（HT_TRENDMODE）值
			"alpha-vantage_HT_DCPERIOD",  // 希尔伯特变换，主周期（HT_DCPERIOD）值
			"alpha-vantage_HT_DCPHASE",   // 希尔伯特变换，主导循环相（HT_DCPHASE）值
			"alpha-vantage_HT_PHASOR",    // 希尔伯特变换，相量分量（HT_PHASOR）值

			"alpha-vantage_ANALYTICS_FIXED_WINDOW",   // 固定窗口上的高级分析
			"alpha-vantage_ANALYTICS_SLIDING_WINDOW": // 滑动窗口上的高级分析
			technicalAnalysisTools = append(technicalAnalysisTools, tool)

		case "alpha-vantage_EARNINGS_CALL_TRANSCRIPT", // 带有LLM情绪的财报电话会议文字记录
			"alpha-vantage_INCOME_STATEMENT", // 年度和季度损益表
			"alpha-vantage_BALANCE_SHEET",    // 年度及季度资产负债表
			"alpha-vantage_CASH_FLOW",        // 年度和季度现金流量表
			"alpha-vantage_EARNINGS":         // 年度及季度盈利数据
			fundamentalAnalysisTools = append(fundamentalAnalysisTools, tool)

		case "alpha-vantage_TIME_SERIES_INTRADAY", // 当前及20+年的历史日内OHLCV数据
			"alpha-vantage_TIME_SERIES_DAILY",            // 覆盖20+年日时序（OHLCV）
			"alpha-vantage_TIME_SERIES_DAILY_ADJUSTED",   // 带拆分/股息事件的日调整OHLCV
			"alpha-vantage_TIME_SERIES_WEEKLY",           // 每周时间序列（每周最后一个交易日）
			"alpha-vantage_TIME_SERIES_WEEKLY_ADJUSTED",  // 带股息的周调整时间序列
			"alpha-vantage_TIME_SERIES_MONTHLY",          // 月度时间序列（每月最后交易日）
			"alpha-vantage_TIME_SERIES_MONTHLY_ADJUSTED", // 带股息的月度调整时间序列
			"alpha-vantage_GLOBAL_QUOTE",                 // 股票代价的最新价格和成交量
			"alpha-vantage_REALTIME_BULK_QUOTES",         // 实时报价最多可达100个符号

			"alpha-vantage_REALTIME_OPTIONS",   // 与希腊的实时美国期权数据
			"alpha-vantage_HISTORICAL_OPTIONS", // 历史期权链15+年

			"alpha-vantage_FX_INTRADAY", // 盘中外汇汇率
			"alpha-vantage_FX_DAILY",    // 每日外汇汇率
			"alpha-vantage_FX_WEEKLY",   // 每周外汇汇率
			"alpha-vantage_FX_MONTHLY",  // 月度外汇汇率

			"alpha-vantage_CURRENCY_EXCHANGE_RATE",    // 数字货币/加密货币之间的汇率
			"alpha-vantage_DIGITAL_CURRENCY_INTRADAY", // 数字货币的日内时间序列
			"alpha-vantage_DIGITAL_CURRENCY_DAILY",    // 数字货币的每日时间序列
			"alpha-vantage_DIGITAL_CURRENCY_WEEKLY",   // 数字货币的每周时间序列
			"alpha-vantage_DIGITAL_CURRENCY_MONTHLY":  // 数字货币的月度时间序列

			comprehensiveAnalysisTools = append(comprehensiveAnalysisTools, tool)
			technicalAnalysisTools = append(technicalAnalysisTools, tool)

		case "alpha-vantage_COMPANY_OVERVIEW", // 公司信息、财务比率和指标
			"alpha-vantage_LISTING_STATUS",       // 股票数据的上市与下架
			"alpha-vantage_EARNINGS_CALENDAR",    // 即将公布财报的财报日历
			"alpha-vantage_IPO_CALENDAR",         // 首次公开募股日程
			"alpha-vantage_NEWS_SENTIMENT",       // 实时及历史市场新闻与情绪
			"alpha-vantage_INSIDER_TRANSACTIONS": // 最新及历史内幕交易
			comprehensiveAnalysisTools = append(comprehensiveAnalysisTools, tool)
			fundamentalAnalysisTools = append(fundamentalAnalysisTools, tool)

		case "alpha-vantage_SYMBOL_SEARCH", // 通过关键词搜索符号
			"alpha-vantage_MARKET_STATUS",      // 全球当前市场状况
			"alpha-vantage_TOP_GAINERS_LOSERS": // 前20名的涨幅、输家和最活跃球员
			comprehensiveAnalysisTools = append(comprehensiveAnalysisTools, tool)
			fundamentalAnalysisTools = append(fundamentalAnalysisTools, tool)
			technicalAnalysisTools = append(technicalAnalysisTools, tool)

		case "alpha-vantage_PING", // 健康检查工具
			"alpha-vantage_ADD_TWO_NUMBERS": // 添加两个数字的示例工具
		}
	}

	return comprehensiveAnalysisTools,
		macroeconomicAnalysisTools,
		fundamentalAnalysisTools,
		technicalAnalysisTools,
		allTools,
		nil
}
