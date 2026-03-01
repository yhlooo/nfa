package commands

import "github.com/nicksnyder/go-i18n/v2/i18n"

var (
	MsgCmdShortDesc = &i18n.Message{
		ID:    "commands.CmdShortDesc",
		Other: "Financial Trading LLM AI Agent. **This is Not Financial Advice.**",
	}

	MsgGlobalOptsVerbosityDesc = &i18n.Message{
		ID:    "commands.GlobalOptsVerbosityDesc",
		Other: "Number for the log level verbosity (0, 1, or 2)",
	}
	MsgGlobalOptsDataRootDesc = &i18n.Message{
		ID:    "commands.GlobalOptsDataRootDesc",
		Other: "Path of data root directory",
	}

	MsgRootOptsMainModelDesc = &i18n.Message{
		ID:    "commands.RootOptsMainModelDesc",
		Other: "Default main model for the current session",
	}
	MsgRootOptsFastModelDesc = &i18n.Message{
		ID:    "commands.RootOptsFastModelDesc",
		Other: "Default fast model for the current session",
	}
	MsgRootOptsPrintAndExitDesc = &i18n.Message{
		ID:    "commands.RootOptsPrintAndExitDesc",
		Other: "Print answer and exit after responding",
	}

	MsgCmdShortDescModels     = &i18n.Message{ID: "commands.CmdShortDescModels", Other: "Manage LLMs used by the agent"}
	MsgCmdShortDescModelsList = &i18n.Message{ID: "commands.CmdShortDescModelsList", Other: "List available models"}

	MsgReasoningTag = &i18n.Message{ID: "commands.ReasoningTag", Other: "Reasoning"}
	MsgVisionTag    = &i18n.Message{ID: "commands.VisionTag", Other: "Vision"}
)
