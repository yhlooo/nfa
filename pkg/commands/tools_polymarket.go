package commands

import (
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/polymarket"
	polymarketui "github.com/yhlooo/nfa/pkg/ui/polymarket"
	polymarketwatcher "github.com/yhlooo/nfa/pkg/ui/polymarketwatcher"
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
	)

	return cmd
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
