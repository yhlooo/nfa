package acputil

import "errors"

var (
	// ErrNotSupported 不支持
	ErrNotSupported = errors.New("NotSupported")
	// ErrSessionNotFound 会话不存在
	ErrSessionNotFound = errors.New("SessionNotFound")
	// ErrInPrompting 对话进行中
	ErrInPrompting = errors.New("InPrompting")
)
