package chat

import "github.com/nicksnyder/go-i18n/v2/i18n"

var (
	MsgNFANote = &i18n.Message{
		ID:    "ui.chat.NFANote",
		Other: "NOTE: Any output should not be construed as financial advice.",
	}

	MsgCmdDescClear = &i18n.Message{
		ID:    "ui.chat.CmdDescClear",
		Other: "Start a fresh conversation",
	}
	MsgCmdDescModel = &i18n.Message{
		ID:    "ui.chat.CmdDescModel",
		Other: "Set the AI model for NFA",
	}
	MsgCmdDescSkills = &i18n.Message{
		ID:    "ui.chat.CmdDescSkills",
		Other: "List loaded skills",
	}
	MsgCmdDescExit = &i18n.Message{
		ID:    "ui.chat.CmdDescExit",
		Other: "Exit the NFA",
	}

	MsgSkillsCount = &i18n.Message{
		ID:    "ui.chat.SkillsCount",
		One:   "{{ .Count }} skill",
		Other: "{{ .Count }} skills",
	}
)
