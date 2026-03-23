package polymarket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

// AuthInfo 认证信息
type AuthInfo struct {
	// 签名者账户地址 (必须)
	//
	// 比如 MetaMask 钱包账户地址，注意不是 PolyMarket 中的代理钱包地址
	Address string
	// 签名者私钥，与 Address 对应 (L1 认证必须)
	PrivateKey string
	// PolyMarket 用户 API Key (L2 认证必须)
	APIKey string
	// PolyMarket 用户 secret (L2 认证必须)
	Secret string
	// PolyMarket 用户 passphrase (L2 认证必须)
	Passphrase string
	// Relayer API Key (Relayer API 必须)
	RelayerAPIKey string
	// Relayer 地址 (Relayer API 必须)
	RelayerAddress string
}

// HasL1Auth 判断是否具有 L1 认证信息
func (auth AuthInfo) HasL1Auth() bool {
	return auth.Address != "" && auth.PrivateKey != ""
}

// L1Sign 生成 L1 签名 (CLOB EIP-712 签名)
func (auth AuthInfo) L1Sign(ts string, nonce *big.Int, message string) (string, error) {
	return "", nil
}

// WithL2Auth 返回加上 L2 的认证信息
func (auth AuthInfo) WithL2Auth(apiKey, secret, passphrase string) AuthInfo {
	newAuth := auth
	newAuth.APIKey = apiKey
	newAuth.Secret = secret
	newAuth.Passphrase = passphrase
	return newAuth
}

// HasL2Auth 判断是否具有 L2 认证信息
func (auth AuthInfo) HasL2Auth() bool {
	return auth.HasL1Auth() && auth.APIKey != "" && auth.Secret != "" && auth.Passphrase != ""
}

// L2Sign 生成 L2 签名 (请求的 HMAC 签名)
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
