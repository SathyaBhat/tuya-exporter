package tuya

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL   string
	accessID  string
	accessKey string
	http      *http.Client
	token     *TokenProvider
}

func NewClient(baseURL, accessID, accessKey string) *Client {
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		accessID:  accessID,
		accessKey: accessKey,
		http:      &http.Client{Timeout: 10 * time.Second},
		token:     &TokenProvider{},
	}
}

func (c *Client) ensureToken(ctx context.Context) (string, error) {
	if tok, ok := c.token.Get(); ok {
		return tok, nil
	}
	return c.refreshToken(ctx)
}

func (c *Client) refreshToken(ctx context.Context) (string, error) {
	path := "/v1.0/token?grant_type=1"
	t := fmt.Sprintf("%d", time.Now().UnixMilli())
	sign := c.calcSign(t, "", "GET", path)

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("client_id", c.accessID)
	req.Header.Set("sign", sign)
	req.Header.Set("t", t)
	req.Header.Set("sign_method", "HMAC-SHA256")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}

	if !tokenResp.Success {
		return "", fmt.Errorf("token request failed: code=%d msg=%s", tokenResp.Code, tokenResp.Msg)
	}

	c.token.Set(tokenResp.Result.AccessToken, tokenResp.Result.ExpireTime)
	slog.Info("tuya token acquired", "expires_in", tokenResp.Result.ExpireTime)
	return tokenResp.Result.AccessToken, nil
}

func (c *Client) GetDeviceStatus(ctx context.Context, deviceID string) ([]StatusItem, error) {
	path := fmt.Sprintf("/v1.0/devices/%s/status", deviceID)

	token, err := c.ensureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	t := fmt.Sprintf("%d", time.Now().UnixMilli())
	sign := c.calcSign(t, token, "GET", path)

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create status request: %w", err)
	}

	req.Header.Set("client_id", c.accessID)
	req.Header.Set("access_token", token)
	req.Header.Set("sign", sign)
	req.Header.Set("t", t)
	req.Header.Set("sign_method", "HMAC-SHA256")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("status request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read status response: %w", err)
	}

	var statusResp DeviceStatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("parse status response: %w", err)
	}

	if !statusResp.Success {
		if statusResp.Code == 1010 || statusResp.Code == 1011 {
			slog.Warn("tuya token expired, refreshing")
			c.token.Invalidate()
		}
		return nil, fmt.Errorf("status request failed: code=%d msg=%s", statusResp.Code, statusResp.Msg)
	}

	return statusResp.Result, nil
}

// GetDeviceLogs fetches historical data point reports (event type 7) for a device.
// startTime and endTime are millisecond timestamps. nextRowKey is used for pagination
// (pass empty string for the first page).
func (c *Client) GetDeviceLogs(ctx context.Context, deviceID string, startTime, endTime int64, codes []string, nextRowKey string) (*DeviceLogsResult, error) {
	path := fmt.Sprintf("/v1.0/devices/%s/logs", deviceID)

	token, err := c.ensureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	// Build query params sorted alphabetically — Tuya signing requires sorted keys
	// and raw (non-percent-encoded) values in the string-to-sign.
	// Sort order: codes, end_time, size, start_row_key, start_time, type
	var queryParts []string
	if len(codes) > 0 {
		queryParts = append(queryParts, "codes="+strings.Join(codes, ","))
	}
	queryParts = append(queryParts,
		fmt.Sprintf("end_time=%d", endTime),
		"size=100",
	)
	if nextRowKey != "" {
		queryParts = append(queryParts, "start_row_key="+nextRowKey)
	}
	queryParts = append(queryParts,
		fmt.Sprintf("start_time=%d", startTime),
		"type=7",
	)
	query := strings.Join(queryParts, "&")
	signPath := path + "?" + query

	t := fmt.Sprintf("%d", time.Now().UnixMilli())
	sign := c.calcSign(t, token, "GET", signPath)

	url := c.baseURL + signPath
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create logs request: %w", err)
	}

	req.Header.Set("client_id", c.accessID)
	req.Header.Set("access_token", token)
	req.Header.Set("sign", sign)
	req.Header.Set("t", t)
	req.Header.Set("sign_method", "HMAC-SHA256")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("logs request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read logs response: %w", err)
	}

	var logsResp DeviceLogsResponse
	if err := json.Unmarshal(body, &logsResp); err != nil {
		return nil, fmt.Errorf("parse logs response: %w", err)
	}

	if !logsResp.Success {
		return nil, fmt.Errorf("logs request failed: code=%d msg=%s", logsResp.Code, logsResp.Msg)
	}

	return &logsResp.Result, nil
}

func (c *Client) GetDeviceInfo(ctx context.Context, deviceID string) (*DeviceInfo, error) {
	path := fmt.Sprintf("/v1.0/devices/%s", deviceID)

	token, err := c.ensureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	t := fmt.Sprintf("%d", time.Now().UnixMilli())
	sign := c.calcSign(t, token, "GET", path)

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create device info request: %w", err)
	}

	req.Header.Set("client_id", c.accessID)
	req.Header.Set("access_token", token)
	req.Header.Set("sign", sign)
	req.Header.Set("t", t)
	req.Header.Set("sign_method", "HMAC-SHA256")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device info request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read device info response: %w", err)
	}

	var infoResp DeviceInfoResponse
	if err := json.Unmarshal(body, &infoResp); err != nil {
		return nil, fmt.Errorf("parse device info response: %w", err)
	}

	if !infoResp.Success {
		return nil, fmt.Errorf("device info request failed: code=%d msg=%s", infoResp.Code, infoResp.Msg)
	}

	return &infoResp.Result, nil
}

func (c *Client) calcSign(t, accessToken, method, path string) string {
	emptyBodyHash := sha256Sum("")
	stringToSign := strings.Join([]string{
		method,
		emptyBodyHash,
		"",
		path,
	}, "\n")

	message := c.accessID + accessToken + t + stringToSign
	return strings.ToUpper(hmacSHA256(message, c.accessKey))
}

func sha256Sum(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func hmacSHA256(message, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
