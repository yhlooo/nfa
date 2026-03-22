package polymarket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
)

// NewCommonClient 创建通用客户端
func NewCommonClient(authInfo AuthInfo) *CommonClient {
	return &CommonClient{
		httpClient: http.DefaultClient,
		authInfo:   authInfo,
	}
}

// CommonClient 通用客户端
type CommonClient struct {
	httpClient *http.Client
	authInfo   AuthInfo
}

// RawRequest 原始请求
type RawRequest struct {
	Method   string
	Endpoint string
	URI      string
	Query    url.Values
	BodyData any
	Headers  http.Header

	WithL2Auth bool
}

// URL 请求地址
func (req RawRequest) URL() string {
	u := req.Endpoint + req.URI
	if len(req.Query) != 0 {
		u += "?" + req.Query.Encode()
	}
	return u
}

// SendRawRequest 发送原始请求，返回响应
func (c *CommonClient) SendRawRequest(ctx context.Context, req *RawRequest) (*http.Response, error) {
	var bodyRaw []byte
	if req.BodyData != nil {
		var err error
		bodyRaw, err = json.Marshal(req.BodyData)
		if err != nil {
			return nil, fmt.Errorf("marshal request body to JSON error: %w", err)
		}
	}

	var body io.Reader
	if bodyRaw != nil {
		body = bytes.NewReader(bodyRaw)
	}

	// 构造请求
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL(), body)
	if err != nil {
		return nil, fmt.Errorf("make request error: %w", err)
	}
	if bodyRaw != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 设置认证头
	if req.WithL2Auth {
		if err := c.authInfo.SetL2AuthHeader(httpReq, bodyRaw); err != nil {
			return nil, fmt.Errorf("set l2 auth header: %w", err)
		}
	}

	// 设置额外请求头
	for k, v := range httpReq.Header {
		httpReq.Header[k] = v
	}

	logger := logr.FromContextOrDiscard(ctx)
	logger.V(1).Info(fmt.Sprintf(
		"send request: %s %s, header: %q",
		httpReq.Method, httpReq.URL, httpReq.Header,
	))

	return c.httpClient.Do(httpReq)
}

// Do 发送请求并处理响应
func (c *CommonClient) Do(ctx context.Context, req *RawRequest, respData any) error {
	resp, err := c.SendRawRequest(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		respContent, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10)) // 读取前 4K
		return fmt.Errorf(
			"%w: status code: %d, body: %s",
			HTTPStatusCodeError(resp.StatusCode), resp.StatusCode, string(respContent),
		)
	}

	if respData == nil {
		// 不需要获取响应内容
		return nil
	}

	// 检查 Content-Type
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		respContent, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10)) // 读取前 4K
		return fmt.Errorf(
			"invalid response content type: %s (expected application/json), body: %s",
			contentType, string(respContent),
		)
	}

	// 读响应，反序列化
	respContent, err := io.ReadAll(io.LimitReader(resp.Body, 500<<20))
	if err != nil {
		return fmt.Errorf("read response body error: %w", err)
	}
	if err := json.Unmarshal(respContent, respData); err != nil {
		return fmt.Errorf("unmarshal response body from JSON error: %w, body: %s", err, string(respContent))
	}

	return nil
}
