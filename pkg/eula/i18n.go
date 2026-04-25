package eula

import "github.com/nicksnyder/go-i18n/v2/i18n"

var (
	// MsgEULAAgreePrompt 询问用户是否同意协议
	MsgEULAAgreePrompt = &i18n.Message{
		ID:    "eula.AgreePrompt",
		Other: "Do you agree to the above terms? (y/n): ",
	}
	// MsgEULADeclined 用户拒绝协议时的提示
	MsgEULADeclined = &i18n.Message{
		ID:    "eula.Declined",
		Other: "You must agree to the End User License Agreement to use this software. Exiting.",
	}
	// MsgEULAInvalidInput 用户输入无效时的提示
	MsgEULAInvalidInput = &i18n.Message{
		ID:    "eula.InvalidInput",
		Other: "Invalid input. Please enter 'y' (yes) or 'n' (no).",
	}
	// MsgEULAUpdated 协议更新时的提示
	MsgEULAUpdated = &i18n.Message{
		ID:    "eula.Updated",
		Other: "The End User License Agreement has been updated. Please review the new terms below:",
	}
)
