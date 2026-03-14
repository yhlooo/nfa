package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/otter"
)

// NewOtterOptions 创建默认 OtterOptions
func NewOtterOptions() OtterOptions {
	return OtterOptions{
		Color:      true,
		Background: true,
		Scale:      1,
	}
}

// OtterOptions otter 子命令选项
type OtterOptions struct {
	// 是否打印彩色图片
	Color bool
	// 是否打印背景
	Background bool
	// 缩放比例
	Scale int
}

// AddPFlags 将选项绑定到命令行参数
func (opts *OtterOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&opts.Color, "color", opts.Color, i18n.T(MsgOtterOptsColorDesc))
	fs.BoolVar(&opts.Background, "bg", opts.Background, i18n.T(MsgOtterOptsBackgroundDesc))
	fs.IntVarP(&opts.Scale, "scale", "x", opts.Scale, i18n.T(MsgOtterOptsScaleDesc))
}

// newOtterCommand 创建 otter 子命令
func newOtterCommand() *cobra.Command {
	opts := NewOtterOptions()
	cmd := &cobra.Command{
		Use:   "otter",
		Short: i18n.T(MsgCmdShortDescOtter),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ret, err := otter.Otter(opts.Color, opts.Background, opts.Scale)
			if err != nil {
				return err
			}
			fmt.Println(ret)
			return nil
		},
	}

	opts.AddPFlags(cmd.Flags())

	return cmd
}
