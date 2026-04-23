package chat

import "github.com/nicksnyder/go-i18n/v2/i18n"

var (
	MsgNFANote = &i18n.Message{
		ID:    "ui.chat.NFANote",
		Other: "NOTE: Any output should not be construed as financial advice.",
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

	MsgTokenUsage    = &i18n.Message{ID: "ui.chat.TokenUsage", Other: "Token Usage:"}
	MsgSkills        = &i18n.Message{ID: "ui.chat.Skills", Other: "Skills"}
	MsgBuiltinSkills = &i18n.Message{ID: "ui.chat.BuiltinSkills", Other: "Builtin skills"}
	MsgLocalSkills   = &i18n.Message{ID: "ui.chat.LocalSkills", Other: "Local skills"}
	MsgSelectModel   = &i18n.Message{ID: "ui.chat.SelectModel", Other: "Select {{ .Type }} model"}
	MsgMultilineMode = &i18n.Message{ID: "ui.chat.MultilineMode", Other: "MULTILINE MODE"}
	MsgTabToToggle   = &i18n.Message{ID: "ui.chat.TabToToggle", Other: "(tab to toggle)"}
	MsgStopReason    = &i18n.Message{ID: "ui.chat.StopReason", Other: "stop reason: {{ .Reason }}"}
	MsgToolCall      = &i18n.Message{ID: "ui.chat.ToolCall", Other: "ToolCall:"}

	MsgSetModel = &i18n.Message{ID: "ui.chat.SetModel", Other: "set {{ .Type }} model:"}

	MsgResumeSession = &i18n.Message{ID: "ui.chat.ResumeSession", Other: "Resume this session with:"}
	MsgResumeCommand = &i18n.Message{ID: "ui.chat.ResumeCommand", Other: "nfa --resume {{ .SessionID }}"}
)
