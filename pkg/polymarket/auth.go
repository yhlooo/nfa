package polymarket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// AuthInfo 认证信息
type AuthInfo struct {
	// 签名者账户地址（比如 MetaMask 钱包账户地址，注意不是 PolyMarket 中的代理钱包地址）
	Address string
	// PolyMarket 用户 API Key
	APIKey string
	// PolyMarket 用户 secret
	Secret string
	// PolyMarket 用户 passphrase
	Passphrase string
	// Relayer API Key
	RelayerAPIKey string
	// Relayer 地址
	RelayerAddress string
}

// L2Sign 生成 L2 签名
func (auth AuthInfo) L2Sign(method, uri, ts string, body []byte) (string, error) {
	secretRaw, err := base64.URLEncoding.DecodeString(auth.Secret)
	if err != nil {
		return "", fmt.Errorf("invalid secret: %w", err)
	}
	h := hmac.New(sha256.New, secretRaw)
	msg := ts + method + uri + string(body)
	h.Write([]byte(msg))
	return base64.URLEncoding.EncodeToString(h.Sum(nil)), nil
}

// SetL2AuthHeader 设置 L2 认证请求头
func (auth AuthInfo) SetL2AuthHeader(req *http.Request, body []byte) error {
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	sign, err := auth.L2Sign(req.Method, req.URL.Path, ts, body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSignError, err)
	}

	req.Header.Set("POLY_ADDRESS", auth.Address)
	req.Header.Set("POLY_SIGNATURE", sign)
	req.Header.Set("POLY_TIMESTAMP", ts)
	req.Header.Set("POLY_API_KEY", auth.APIKey)
	req.Header.Set("POLY_PASSPHRASE", auth.Passphrase)

	return nil
}
