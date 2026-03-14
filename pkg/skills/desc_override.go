package skills

import (
	"context"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	i18nutil "github.com/yhlooo/nfa/pkg/i18n"
)

var (
	MsgCmdShortDesc = &i18n.Message{
		ID:    "skills.ShortTermTrendForecastDesc",
		Other: "Analyze and predict short-term stock trends (within days, week, or month).",
	}
)

// BuiltinSkillsDescOverride 内置 Skill 描述覆盖
func BuiltinSkillsDescOverride(ctx context.Context) map[string]string {
	return map[string]string{
		"short-term-trend-forecast": i18nutil.TContext(ctx, MsgCmdShortDesc),
	}
}
