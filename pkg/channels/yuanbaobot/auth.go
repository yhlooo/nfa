package yuanbaobot

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

const (
	// SignTokenPath Sign-Token API 路径
	SignTokenPath = "/api/v5/robotLogic/sign-token"
	// TokenRefreshMargin Token 刷新提前量
	TokenRefreshMargin = 5 * time.Minute
)

// SignTokenRequest Sign-Token 请求
type SignTokenRequest struct {
	AppKey    string `json:"app_key"`
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
	Timestamp string `json:"timestamp"`
}

// SignTokenResponse Sign-Token 响应
type SignTokenResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		BotID      string `json:"bot_id"`
		Token      string `json:"token"`
		Duration   int64  `json:"duration"` // 有效期（秒）
		Product    string `json:"product"`
		Source     string `json:"source"`
		CreateType int    `json:"create_type"` // 1=一键创建, 2=关联创建
	} `json:"data"`
}

// TokenCache Token 缓存
type TokenCache struct {
	Token     string
	BotID     string
	Source    string
	ExpiresAt time.Time
	Duration  int64
}

// AuthManager 认证管理器
type AuthManager struct {
	AppKey    string
	AppSecret string
	APIDomain string

	lock  sync.RWMutex
	cache *TokenCache

	httpClient *http.Client
	logger     logr.Logger
}

// NewAuthManager 创建认证管理器
func NewAuthManager(appKey, appSecret, apiDomain string, logger logr.Logger) *AuthManager {
	return &AuthManager{
		AppKey:     appKey,
		AppSecret:  appSecret,
		APIDomain:  apiDomain,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

// computeSignature 计算 HMAC-SHA256 签名
// plain = nonce + timestamp + appKey + appSecret
func computeSignature(appSecret, nonce, timestamp, appKey string) string {
	plain := nonce + timestamp + appKey + appSecret
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(plain))
	return hex.EncodeToString(mac.Sum(nil))
}

// generateNonce 生成随机 nonce
func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SignToken 获取 Token（优先使用缓存）
func (a *AuthManager) SignToken(ctx context.Context) (*TokenCache, error) {
	a.lock.RLock()
	if a.cache != nil && time.Now().Before(a.cache.ExpiresAt) {
		cache := a.cache
		a.lock.RUnlock()
		return cache, nil
	}
	a.lock.RUnlock()

	return a.RefreshToken(ctx)
}

// RefreshToken 强制刷新 Token
func (a *AuthManager) RefreshToken(ctx context.Context) (*TokenCache, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("generate nonce error: %w", err)
	}

	// 使用北京时间（+8）
	timestamp := time.Now().In(time.FixedZone("CST", 8*3600)).Format(time.RFC3339)
	signature := computeSignature(a.AppSecret, nonce, timestamp, a.AppKey)

	reqBody := SignTokenRequest{
		AppKey:    a.AppKey,
		Nonce:     nonce,
		Signature: signature,
		Timestamp: timestamp,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal sign token request error: %w", err)
	}

	url := fmt.Sprintf("%s%s", a.APIDomain, SignTokenPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create sign token request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sign token request error: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if err != nil {
		return nil, fmt.Errorf("read sign token response error: %w", err)
	}

	var signResp SignTokenResponse
	if err := json.Unmarshal(raw, &signResp); err != nil {
		return nil, fmt.Errorf("unmarshal sign token response error: %w, body: %s", err, string(raw))
	}

	if signResp.Code != 0 {
		return nil, fmt.Errorf("sign token failed: code %d, msg: %s", signResp.Code, signResp.Msg)
	}

	duration := time.Duration(signResp.Data.Duration) * time.Second
	cache := &TokenCache{
		Token:     signResp.Data.Token,
		BotID:     signResp.Data.BotID,
		Source:    signResp.Data.Source,
		ExpiresAt: time.Now().Add(duration - TokenRefreshMargin),
		Duration:  signResp.Data.Duration,
	}
	a.cache = cache

	a.logger.Info("sign token success", "botID", cache.BotID, "duration", signResp.Data.Duration)
	return cache, nil
}
