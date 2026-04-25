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
	MsgGlobalOptsLangDesc = &i18n.Message{
		ID:    "commands.GlobalOptsLangDesc",
		Other: "The language used in UI (en or zh)",
	}

	MsgRootOptsModelDesc = &i18n.Message{
		ID:    "commands.RootOptsModelDesc",
		Other: "Primary model for the current session",
	}
	MsgRootOptsLightModelDesc = &i18n.Message{
		ID:    "commands.RootOptsLightModelDesc",
		Other: "Light model for the current session",
	}
	MsgRootOptsVisionModelDesc = &i18n.Message{
		ID:    "commands.RootOptsVisionModelDesc",
		Other: "Vision model for the current session",
	}
	MsgRootOptsPrintAndExitDesc = &i18n.Message{
		ID:    "commands.RootOptsPrintAndExitDesc",
		Other: "Print answer and exit after responding",
	}
	MsgRootOptsResumeDesc = &i18n.Message{
		ID:    "commands.RootOptsResumeDesc",
		Other: "Resume a previous session by session ID",
	}

	MsgCmdShortDescModels     = &i18n.Message{ID: "commands.CmdShortDescModels", Other: "Manage LLMs used by the agent"}
	MsgCmdShortDescModelsList = &i18n.Message{ID: "commands.CmdShortDescModelsList", Other: "List available models"}

	MsgModelNameTag    = &i18n.Message{ID: "commands.ModelNameTag", Other: "Name"}
	MsgReasoningTag    = &i18n.Message{ID: "commands.ReasoningTag", Other: "Reasoning"}
	MsgVisionTag       = &i18n.Message{ID: "commands.VisionTag", Other: "Vision"}
	MsgModelContextTag = &i18n.Message{ID: "commands.ModelContextTag", Other: "Context"}

	MsgScoreTag = &i18n.Message{ID: "commands.ScoreTag", Other: "Score"}

	MsgCmdShortDescOtter = &i18n.Message{ID: "commands.CmdShortDescOtter", Other: "Print Otter image"}

	MsgOtterOptsColorDesc      = &i18n.Message{ID: "commands.OtterOptsColorDesc", Other: "Print with color"}
	MsgOtterOptsBackgroundDesc = &i18n.Message{ID: "commands.OtterOptsBackgroundDesc", Other: "Print with background"}
	MsgOtterOptsScaleDesc      = &i18n.Message{ID: "commands.OtterOptsScaleDesc", Other: "Scaling factor"}

	MsgCmdShortDescVersion         = &i18n.Message{ID: "commands.CmdShortDescVersion", Other: "Print the version information"}
	MsgVersionOptsOutputFormatDesc = &i18n.Message{ID: "commands.VersionOptsOutputFormatDesc", Other: "Output format. One of (json)"}
)
