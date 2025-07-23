package alarm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPRequest HTTP 请求配置
type HTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

// HTTPResponse HTTP 响应结构
type HTTPResponse struct {
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
}

// SendHTTPRequest 发送 HTTP 请求的统一方法
func SendHTTPRequest(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error) {
	// 获取统一的 HTTP 客户端
	client := getHTTPClient()
	if client == nil {
		return nil, fmt.Errorf("http client not initialized")
	}

	// 创建请求
	var httpReq *http.Request
	var err error

	if req.Body != nil {
		httpReq, err = http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	} else {
		httpReq, err = http.NewRequestWithContext(ctx, req.Method, req.URL, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// 设置请求头
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 发送请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}

// SendHTTPRequestWithRetry 带重试的 HTTP 请求
func SendHTTPRequestWithRetry(ctx context.Context, req *HTTPRequest, maxRetry int) error {
	var lastErr error
	for i := 0; i < maxRetry; i++ {
		resp, err := SendHTTPRequest(ctx, req)
		if err != nil {
			lastErr = err
			if i < maxRetry-1 {
				// 等待一段时间后重试
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second * time.Duration(i+1)):
					continue
				}
			}
			continue
		}

		// 检查状态码
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(resp.Body))
		if i < maxRetry-1 {
			// 等待一段时间后重试
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second * time.Duration(i+1)):
				continue
			}
		}
	}

	return lastErr
}
