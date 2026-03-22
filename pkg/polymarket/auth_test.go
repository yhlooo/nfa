package polymarket

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthInfo_L2Sign 测试 AuthInfo.L2Sign 方法
func TestAuthInfo_L2Sign(t *testing.T) {
	a := assert.New(t)

	auth := AuthInfo{
		Secret: "iO5A44JWuUPsfZ0pzKEX11ez4eLg_yvNLnuXwm96iD8=", // 随机生成的密钥，无任何实际作用
	}
	s, err := auth.L2Sign(http.MethodGet, "/test", "1774178860", []byte(`{"key":"value"}`))
	a.NoError(err)
	a.Equal("_-XiITiB2hE2ekY4d_urRAa3GOZXVYiysVvmAyMYxw8=", s)
}

// TestAuthInfo_SetL2AuthHeader 测试 AuthInfo.SetL2AuthHeader 方法
func TestAuthInfo_SetL2AuthHeader(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	// 构造测试请求
	req, err := http.NewRequest(http.MethodGet, "/test?k=v", nil)
	r.NoError(err)

	// 构造测试认证信息
	// 随机生成的认证信息，无任何实际作用
	auth := AuthInfo{
		Address:    "0x1234567890123456789012345678901234567890",
		APIKey:     "12345678-1234-1234-1234-123456789012",
		Secret:     "iO5A44JWuUPsfZ0pzKEX11ez4eLg_yvNLnuXwm96iD8=",
		Passphrase: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	}

	r.NoError(auth.SetL2AuthHeader(req, []byte(`{"key":"value"}`)))
	a.Equal("0x1234567890123456789012345678901234567890", req.Header.Get("POLY_ADDRESS"))
	a.NotEmpty(req.Header.Get("POLY_SIGNATURE"))
	a.Len(req.Header.Get("POLY_TIMESTAMP"), 10)
	a.Equal("12345678-1234-1234-1234-123456789012", req.Header.Get("POLY_API_KEY"))
	a.Equal("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", req.Header.Get("POLY_PASSPHRASE"))
}
