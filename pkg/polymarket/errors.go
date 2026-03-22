package polymarket

import (
	"errors"
	"net/http"
)

var (
	// ErrBadRequest 错误请求
	ErrBadRequest = errors.New("BadRequest")
	// ErrUnauthorized 未认证错误
	ErrUnauthorized = errors.New("Unauthorized")
	// ErrNotFound 资源未找到错误
	ErrNotFound = errors.New("NotFound")
	// ErrUnexpectedStatusCode 其它非预期状态码
	ErrUnexpectedStatusCode = errors.New("UnexpectedStatusCode")
	// ErrSignError 签名错误
	ErrSignError = errors.New("SignError")
)

// HTTPStatusCodeError HTTP 状态码错误
func HTTPStatusCodeError(statusCode int) error {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrNotFound
	default:
		return ErrUnexpectedStatusCode
	}
}
