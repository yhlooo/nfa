package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/polymarket"
	trading2 "github.com/yhlooo/nfa/pkg/polymarket/trading"
	polymarketui "github.com/yhlooo/nfa/pkg/ui/polymarket"
	tradingui "github.com/yhlooo/nfa/pkg/ui/polymarkettrading"
	"github.com/yhlooo/nfa/pkg/ui/polymarketwatcher"
)

// newToolsCommand 创建 tools 子命令
func newToolsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: i18n.T(MsgCmdShortDescTools),
	}

	cmd.AddCommand(
		newPolyMarketCommand(),
	)

	return cmd
}

// newPolyMarketCommand 创建 polymarket 子命令
func newPolyMarketCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "polymarket",
		Short: i18n.T(MsgCmdShortDescToolsPolyMarket),
		// 默认行为：进入交互式浏览器模式
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// 创建客户端（无需认证）
			client := polymarket.NewClient(polymarket.AuthInfo{})

			// 启动浏览器
			browser := polymarketui.NewBrowser(polymarketui.Options{
				Client: client,
			})
			return browser.Run(ctx)
		},
	}

	cmd.AddCommand(
		newPolyMarketWatchCommand(),
		newPolyMarketTradeCommand(),
	)

	return cmd
}

// TradeOptions trade 命令选项
type TradeOptions struct {
	DryRun     bool
	MarketSlug string
	Strategy   string
	Multiplier float64
	Interval   time.Duration
}

// NewTradeOptions 创建 trade 命令选项（默认值）
func NewTradeOptions() TradeOptions {
	return TradeOptions{
		DryRun:     true,
		Multiplier: 1,
		Interval:   5 * time.Second,
	}
}

// AddPFlags 绑定命令行 flags
func (o *TradeOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.DryRun, "dry-run", "d", o.DryRun, "dry-run mode (simulated trading)")
	fs.StringVarP(&o.MarketSlug, "market", "m", o.MarketSlug, "market slug (required)")
	fs.StringVarP(&o.Strategy, "strategy", "s", o.Strategy, "strategy name (required)")
	fs.Float64VarP(&o.Multiplier, "multiplier", "x", o.Multiplier, "trade multiplier")
	fs.DurationVarP(&o.Interval, "interval", "i", o.Interval, "strategy execution interval")
}

// newPolyMarketTradeCommand 创建 polymarket trade 子命令
func newPolyMarketTradeCommand() *cobra.Command {
	opts := NewTradeOptions()

	cmd := &cobra.Command{
		Use:   "trade",
		Short: i18n.T(MsgCmdShortDescToolsPolyMarketTrade),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx)

			// 验证必需参数
			if opts.MarketSlug == "" {
				return fmt.Errorf("market slug is required, use -m <slug>")
			}
			if opts.Strategy == "" {
				return fmt.Errorf("strategy name is required, use -s <strategy>")
			}

			// 创建客户端（无需认证）
			client := polymarket.NewClient(polymarket.AuthInfo{})

			// 获取市场信息
			logger.V(1).Info(fmt.Sprintf("fetching market info for slug: %s", opts.MarketSlug))
			market, err := client.GetMarketBySlug(ctx, opts.MarketSlug)
			if err != nil {
				return fmt.Errorf("get market by slug error: %w", err)
			}

			// 获取策略
			var strategy trading2.Strategy
			switch opts.Strategy {
			case "simple":
				strategy = trading2.NewSimpleStrategy()
			case "rand":
				strategy = trading2.NewRandomStrategy()
			default:
				return fmt.Errorf("unknown strategy: %s (available: simple, rand)", opts.Strategy)
			}

			// 创建执行器
			executor := trading2.NewExecutor(client, market, strategy, trading2.ExecutorOptions{
				DryRun:     opts.DryRun,
				Multiplier: opts.Multiplier,
				Interval:   opts.Interval,
			})

			// 启动执行器
			if err := executor.Run(ctx); err != nil {
				return fmt.Errorf("start executor error: %w", err)
			}
			defer func() { _ = executor.Stop() }()

			// 创建并运行 UI
			page := tradingui.NewPage(executor, strategy, opts.DryRun, opts.Multiplier, opts.MarketSlug)
			return runTradingUI(ctx, page)
		},
	}

	opts.AddPFlags(cmd.Flags())

	return cmd
}

// runTradingUI 运行交易 UI
func runTradingUI(ctx context.Context, page *tradingui.Page) error {
	prog := tea.NewProgram(page, tea.WithContext(ctx))
	_, err := prog.Run()
	return err
}

// newPolyMarketWatchCommand 创建 polymarket watch 子命令
func newPolyMarketWatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch [MARKET_SLUG]",
		Short: i18n.T(MsgCmdShortDescToolsPolyMarketWatch),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx)
			slug := args[0]

			// 创建客户端（无需认证）
			client := polymarket.NewClient(polymarket.AuthInfo{})

			// 获取市场信息
			logger.V(1).Info(fmt.Sprintf("fetching market info for slug: %s", slug))
			market, err := client.GetMarketBySlug(ctx, slug)
			if err != nil {
				return fmt.Errorf("get market by slug error: %w", err)
			}

			// 解析 clobTokenIds 和 outcomes
			var assetIDs []string
			if err := json.Unmarshal([]byte(market.ClobTokenIDs), &assetIDs); err != nil {
				return fmt.Errorf("parse clob token ids error: %w", err)
			}

			var outcomeNames []string
			if err := json.Unmarshal([]byte(market.Outcomes), &outcomeNames); err != nil {
				return fmt.Errorf("parse outcomes error: %w", err)
			}

			logger.V(1).Info(fmt.Sprintf("market: %s, asset IDs: %v, outcomes: %v",
				market.Question, assetIDs, outcomeNames))

			// 创建监听器
			watcher := polymarket.NewWatcher(client, market, assetIDs)
			if err := watcher.Start(ctx); err != nil {
				return fmt.Errorf("start watcher error: %w", err)
			}
			defer func() { _ = watcher.Stop() }()

			// 启动 UI
			ui := polymarketwatcher.NewUI(polymarketwatcher.Options{
				Market:       market,
				AssetIDs:     assetIDs,
				OutcomeNames: outcomeNames,
				Watcher:      watcher,
			})

			return ui.Run(ctx)
		},
	}

	return cmd
}
