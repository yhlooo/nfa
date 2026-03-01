package commands

import (
	"crypto/tls"
	"fmt"
	"io"
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
	"github.com/yhlooo/nfa/pkg/configs"
	"github.com/yhlooo/nfa/pkg/i18n"
	uitty "github.com/yhlooo/nfa/pkg/ui/chat"
	"github.com/yhlooo/nfa/pkg/version"
)

// NewGlobalOptions 创建默认 GlobalOptions
func NewGlobalOptions() GlobalOptions {
	homeDir, _ := os.UserHomeDir()
	return GlobalOptions{
		Verbosity: 0,
		DataRoot:  filepath.Join(homeDir, ".nfa"),
	}
}

// GlobalOptions 全局选项
type GlobalOptions struct {
	// 日志数量级别（ 0 / 1 / 2 ）
	Verbosity uint32
	// 数据存储根目录
	DataRoot string
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
}

// NewOptions 创建默认 Options
func NewOptions() Options {
	return Options{
		MainModel:    "",
		FastModel:    "",
		PrintAndExit: false,
	}
}

// Options 运行选项
type Options struct {
	MainModel    string
	FastModel    string
	PrintAndExit bool
}

// AddPFlags 将选项绑定到命令行参数
func (o *Options) AddPFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.MainModel, "model", o.MainModel, i18n.T(MsgRootOptsMainModelDesc))
	fs.StringVar(&o.FastModel, "fast-model", o.FastModel, i18n.T(MsgRootOptsFastModelDesc))
	fs.BoolVarP(&o.PrintAndExit, "print", "p", o.PrintAndExit, i18n.T(MsgRootOptsPrintAndExitDesc))
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
			ctx = i18n.ContextWithLocalizer(ctx, i18n.NewLocalizer(cfg.Language, i18n.GetEnvLanguage()))

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
			if opts.MainModel != "" {
				m.Main = opts.MainModel
			}
			if opts.FastModel != "" {
				m.Fast = opts.FastModel
			}

			agent := agents.NewNFA(agents.Options{
				Logger:         logger,
				ModelProviders: cfg.ModelProviders,
				DataProviders:  cfg.DataProviders,
				DefaultModels:  m,
				DataRoot:       globalOpts.DataRoot,
			})

			agentIn, clientOut := io.Pipe()
			clientIn, agentOut := io.Pipe()
			defer func() {
				_ = clientOut.Close()
				_ = agentIn.Close()
				_ = agentOut.Close()
				_ = clientIn.Close()
			}()

			if err := agent.Connect(agentIn, agentOut); err != nil {
				return fmt.Errorf("create agent side connection error: %w", err)
			}

			ctx, cancel := chromedp.NewContext(ctx)
			defer cancel()

			// 处理三种模式
			var initialPrompt string
			if len(args) > 0 {
				initialPrompt = args[0]
			}

			return uitty.NewChatUI(uitty.Options{
				AgentClientIn:         clientIn,
				AgentClientOut:        clientOut,
				InitialPrompt:         initialPrompt,
				AutoExitAfterResponse: opts.PrintAndExit,
			}).Run(ctx)
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
		newModelsCommand(),
		newInternalToolsCommand(),
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
