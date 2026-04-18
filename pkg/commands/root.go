package commands

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bombsimon/logrusr/v4"
	"github.com/chromedp/chromedp"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/yhlooo/nfa/pkg/agents"
	uitty "github.com/yhlooo/nfa/pkg/apps/chat"
	"github.com/yhlooo/nfa/pkg/channels"
	"github.com/yhlooo/nfa/pkg/channels/wecomaibot"
	"github.com/yhlooo/nfa/pkg/channels/yuanbaobot"
	"github.com/yhlooo/nfa/pkg/configs"
	"github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/version"
)

// NewGlobalOptions 创建默认 GlobalOptions
func NewGlobalOptions() GlobalOptions {
	homeDir, _ := os.UserHomeDir()
	return GlobalOptions{
		Verbosity: 0,
		DataRoot:  filepath.Join(homeDir, ".nfa"),
		Language:  "",
	}
}

// GlobalOptions 全局选项
type GlobalOptions struct {
	// 日志数量级别（ 0 / 1 / 2 ）
	Verbosity uint32
	// 数据存储根目录
	DataRoot string
	// 语言
	Language string
}

// Validate 校验选项是否合法
func (o *GlobalOptions) Validate() error {
	if o.Verbosity > 2 {
		return fmt.Errorf("invalid log verbosity: %d (expected: 0, 1 or 2)", o.Verbosity)
	}
	return nil
}

// AddPFlags 将选项绑定到命令行参数
func (o *GlobalOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.Uint32VarP(&o.Verbosity, "verbose", "v", o.Verbosity, i18n.T(MsgGlobalOptsVerbosityDesc))
	fs.StringVar(&o.DataRoot, "data-root", o.DataRoot, i18n.T(MsgGlobalOptsDataRootDesc))
	fs.StringVar(&o.Language, "lang", o.Language, i18n.T(MsgGlobalOptsLangDesc))
}

// NewOptions 创建默认 Options
func NewOptions() Options {
	return Options{
		Model:        "",
		LightModel:   "",
		VisionModel:  "",
		PrintAndExit: false,
		Resume:       "",
	}
}

// Options 运行选项
type Options struct {
	Model        string
	LightModel   string
	VisionModel  string
	PrintAndExit bool
	Resume       string
}

// AddPFlags 将选项绑定到命令行参数
func (o *Options) AddPFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Model, "model", o.Model, i18n.T(MsgRootOptsModelDesc))
	fs.StringVar(&o.LightModel, "light-model", o.LightModel, i18n.T(MsgRootOptsLightModelDesc))
	fs.StringVar(&o.VisionModel, "vision-model", o.VisionModel, i18n.T(MsgRootOptsVisionModelDesc))
	fs.BoolVarP(&o.PrintAndExit, "print", "p", o.PrintAndExit, i18n.T(MsgRootOptsPrintAndExitDesc))
	fs.StringVar(&o.Resume, "resume", o.Resume, i18n.T(MsgRootOptsResumeDesc))
}

// NewCommand 创建根命令
func NewCommand(name string) *cobra.Command {
	globalOpts := NewGlobalOptions()
	opts := NewOptions()

	var keylog *os.File
	cmd := &cobra.Command{
		Use:           fmt.Sprintf("%s [PROMPT]", name),
		Short:         i18n.T(MsgCmdShortDesc),
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.Version,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := globalOpts.Validate(); err != nil {
				return err
			}

			// 创建日志目录
			if err := os.MkdirAll(globalOpts.DataRoot, 0o755); err != nil {
				return fmt.Errorf("create log directory %q error: %w", globalOpts.DataRoot, err)
			}

			// 初始化 logger
			logrusLogger := logrus.New()
			logrusLogger.SetOutput(&lumberjack.Logger{
				Filename:   filepath.Join(globalOpts.DataRoot, "nfa.log"),
				MaxSize:    500, // MB
				MaxBackups: 3,
				MaxAge:     28, // 天
			})
			switch globalOpts.Verbosity {
			case 0:
				logrusLogger.Level = logrus.InfoLevel
			case 1:
				logrusLogger.Level = logrus.DebugLevel
			default:
				logrusLogger.Level = logrus.TraceLevel
			}
			logger := logrusr.New(logrusLogger)
			ctx = logr.NewContext(ctx, logger)

			// 加载配置
			cfgPath := filepath.Join(globalOpts.DataRoot, "nfa.json")
			cfg, err := configs.LoadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("load config %q error: %w", cfgPath, err)
			}
			ctx = configs.ContextWithConfig(ctx, cfg, cfgPath)

			// 设置本地化器
			ctx = i18n.ContextWithLocalizer(ctx, i18n.NewLocalizer(globalOpts.Language, cfg.Language, i18n.GetEnvLanguage()))

			keylog, err = setKeyLog()
			if err != nil {
				return fmt.Errorf("set tls key log error: %w", err)
			}

			cmd.SetContext(ctx)

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg := configs.ConfigFromContext(ctx)
			logger := logr.FromContextOrDiscard(ctx)

			m := cfg.DefaultModels
			if opts.Model != "" {
				m.Primary = opts.Model
			}
			if opts.LightModel != "" {
				m.Light = opts.LightModel
			}
			if opts.VisionModel != "" {
				m.Vision = opts.VisionModel
			}

			// 创建 Agent
			agent := agents.NewNFA(agents.Options{
				Logger:         logger,
				Localizer:      i18n.LocalizerFromContext(ctx),
				ModelProviders: cfg.ModelProviders,
				DataProviders:  cfg.DataProviders,
				DefaultModels:  m,
				DataRoot:       globalOpts.DataRoot,
			})

			// 连接信道
			var chs []channels.Channel
			if cfg.Channels.Enabled {
				for _, chOpts := range cfg.Channels.Channels {
					switch {
					case chOpts.WeComAIBot != nil:
						ch := &wecomaibot.WeComAIBot{
							BotID:  chOpts.WeComAIBot.BotID,
							Secret: chOpts.WeComAIBot.Secret,
							URL:    chOpts.WeComAIBot.URL,
						}
						ch.Start(ctx)
						chs = append(chs, ch)
					case chOpts.YuanbaoBot != nil:
						ch := &yuanbaobot.YuanbaoBot{
							AppKey:       chOpts.YuanbaoBot.AppID,
							AppSecret:    chOpts.YuanbaoBot.AppSecret,
							BaseURL:      chOpts.YuanbaoBot.BaseURL,
							WebSocketURL: chOpts.YuanbaoBot.WebSocketURL,
						}
						ch.Start(ctx)
						chs = append(chs, ch)
					}
				}
			}

			// 创建应用
			var initialPrompt string
			if len(args) > 0 {
				initialPrompt = args[0]
			}
			app := uitty.NewChat(uitty.Options{
				Agent:                 agent,
				InitialPrompt:         initialPrompt,
				AutoExitAfterResponse: opts.PrintAndExit,
				ResumeSessionID:       opts.Resume,
				Channels:              chs,
			})
			agent.SetClient(app)

			ctx, cancel := chromedp.NewContext(ctx)
			defer cancel()

			// 开始运行
			return app.Run(ctx)
		},

		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if keylog != nil {
				_ = keylog.Close()
			}
			return nil
		},
	}

	globalOpts.AddPFlags(cmd.PersistentFlags())
	opts.AddPFlags(cmd.Flags())

	cmd.AddCommand(
		newOtterCommand(),
		newModelsCommand(),
		newInternalToolsCommand(),
		newVersionCommand(),
	)

	return cmd
}

// setKeyLog 设置 TLS keylog
func setKeyLog() (*os.File, error) {
	keylog := os.Getenv("SSLKEYLOGFILE")
	if keylog == "" {
		return nil, nil
	}

	if err := os.MkdirAll(filepath.Dir(keylog), 0o755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(keylog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	// 设置输出 keylog 文件
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,

			TLSClientConfig: &tls.Config{KeyLogWriter: f},
		},
	}

	return f, nil
}
