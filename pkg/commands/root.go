package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/yhlooo/nfa/pkg/agents"
	uitty "github.com/yhlooo/nfa/pkg/ui/tty"
	"github.com/yhlooo/nfa/pkg/version"
)

// NewGlobalOptions 创建一个默认 GlobalOptions
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

// NewCommand 创建根命令
func NewCommand(name string) *cobra.Command {
	globalOpts := NewGlobalOptions()

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

			// 创建日志目录
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get user home directory error: %w", err)
			}
			logPath := filepath.Join(home, ".nfa", "nfa.log")
			if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
				return fmt.Errorf("create log directory %q error: %w", filepath.Dir(logPath), err)
			}

			// 初始化 logger
			logrusLogger := logrus.New()
			logrusLogger.SetOutput(&lumberjack.Logger{
				Filename:   logPath,
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

			// 注入 logger 到上下文
			cmd.SetContext(logr.NewContext(ctx, logger))

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			agent := agents.NewNFA(agents.Options{})

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
			}).Run(cmd.Context())
		},
	}

	globalOpts.AddPFlags(cmd.PersistentFlags())

	return cmd
}
