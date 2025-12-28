package commands

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/configs"
	uitty "github.com/yhlooo/nfa/pkg/ui/tty"
	"github.com/yhlooo/nfa/pkg/version"
)

// NewGlobalOptions 创建默认 GlobalOptions
func NewGlobalOptions() GlobalOptions {
	return GlobalOptions{
		Verbosity: 0,
	}
}

// GlobalOptions 全局选项
type GlobalOptions struct {
	// 日志数量级别（ 0 / 1 / 2 ）
	Verbosity uint32
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
	fs.Uint32VarP(&o.Verbosity, "verbose", "v", o.Verbosity, "Number for the log level verbosity (0, 1, or 2)")
}

// NewOptions 创建默认 Options
func NewOptions() Options {
	return Options{
		DefaultModel: "",
	}
}

// Options 运行选项
type Options struct {
	DefaultModel string
	SingleAgent  bool
}

// AddPFlags 将选项绑定到命令行参数
func (o *Options) AddPFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.DefaultModel, "model", o.DefaultModel, "Default model for the current session")
	fs.BoolVar(&o.SingleAgent, "single", o.SingleAgent, "Run in single agent mode")
}

// NewCommand 创建根命令
func NewCommand(name string) *cobra.Command {
	globalOpts := NewGlobalOptions()
	opts := NewOptions()

	var keylog *os.File
	cmd := &cobra.Command{
		Use:           fmt.Sprintf("%s", name),
		Short:         "Financial Trading LLM AI Agent. **This is Not Financial Advice.**",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.Version,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := globalOpts.Validate(); err != nil {
				return err
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get user home directory error: %w", err)
			}
			dataPath := filepath.Join(home, ".nfa")

			// 创建日志目录
			if err := os.MkdirAll(dataPath, 0o755); err != nil {
				return fmt.Errorf("create log directory %q error: %w", dataPath, err)
			}

			// 初始化 logger
			logrusLogger := logrus.New()
			logrusLogger.SetOutput(&lumberjack.Logger{
				Filename:   filepath.Join(dataPath, "nfa.log"),
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
			cfgPath := filepath.Join(dataPath, "nfa.json")
			cfg, err := configs.LoadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("load config %q error: %w", cfgPath, err)
			}
			ctx = NewContextWithConfig(ctx, cfg)

			keylog, err = setKeyLog()
			if err != nil {
				return fmt.Errorf("set tls key log error: %w", err)
			}

			cmd.SetContext(ctx)

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ConfigFromContext(ctx)
			logger := logr.FromContextOrDiscard(ctx)

			defaultModel := opts.DefaultModel
			if defaultModel == "" {
				defaultModel = cfg.DefaultModel
			}

			agent := agents.NewNFA(agents.Options{
				Logger:         logger,
				ModelProviders: cfg.ModelProviders,
				DataProviders:  cfg.DataProviders,
				DefaultModel:   defaultModel,
				SingleAgent:    opts.SingleAgent,
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

			return uitty.NewChatUI(uitty.Options{
				AgentClientIn:  clientIn,
				AgentClientOut: clientOut,
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
	)

	return cmd
}

// configContextKey 上下文中存放配置信息的 key
type configContextKey struct{}

// NewContextWithConfig 创建携带配置信息的上下文
func NewContextWithConfig(parent context.Context, config configs.Config) context.Context {
	return context.WithValue(parent, configContextKey{}, config)
}

// ConfigFromContext 从上下文获取配置信息
func ConfigFromContext(ctx context.Context) configs.Config {
	cfg, ok := ctx.Value(configContextKey{}).(configs.Config)
	if !ok {
		return configs.Config{}
	}
	return cfg
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
