package yuanbaobot

import "errors"

var (
	// ErrAuthFailed 认证失败
	ErrAuthFailed = errors.New("auth failed")
	// ErrNotConnected 未连接
	ErrNotConnected = errors.New("not connected")
)
