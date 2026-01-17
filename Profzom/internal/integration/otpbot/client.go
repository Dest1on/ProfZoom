package otpbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client interface {
	GetTelegramStatus(ctx context.Context, phone string) (Status, error)
	RegisterLinkToken(ctx context.Context, userID, token string) error
	SendOTP(ctx context.Context, phone, otpCode string) error
}

type HTTPClient struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func NewClient(baseURL, internalKey string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPClient{
		baseURL:     strings.TrimRight(baseURL, "/"),
		internalKey: internalKey,
		httpClient:  httpClient,
	}
}

func (c *HTTPClient) GetTelegramStatus(ctx context.Context, phone string) (Status, error) {
	if phone == "" {
		return Status{}, fmt.Errorf("phone is required")
	}
	endpoint := c.baseURL + "/telegram/status"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Status{}, fmt.Errorf("create status request: %w", err)
	}
	if c.internalKey == "" {
		return Status{}, ErrUnauthorized
	}
	query := req.URL.Query()
	query.Set("phone", phone)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("X-Internal-Key", c.internalKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Status{}, fmt.Errorf("send status request: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		// см. ниже
	case http.StatusBadRequest:
		body := readBodySnippet(resp.Body)
		return Status{}, fmt.Errorf("%w: status=%d body=%s", ErrBadRequest, resp.StatusCode, body)
	case http.StatusUnauthorized:
		return Status{}, ErrUnauthorized
	case http.StatusTooManyRequests:
		return Status{}, ErrRateLimited
	default:
		body := readBodySnippet(resp.Body)
		return Status{}, fmt.Errorf("%w: status=%d body=%s", ErrDeliveryFailed, resp.StatusCode, body)
	}
	var status Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return Status{}, fmt.Errorf("%w: decode status response: %v", ErrDeliveryFailed, err)
	}
	return status, nil
}

func (c *HTTPClient) RegisterLinkToken(ctx context.Context, userID, token string) error {
	if userID == "" {
		return fmt.Errorf("%w: user_id is required", ErrDeliveryFailed)
	}
	if token == "" {
		return fmt.Errorf("%w: token is required", ErrDeliveryFailed)
	}
	payload := struct {
		UserID string `json:"user_id"`
		Token  string `json:"token"`
	}{
		UserID: userID,
		Token:  token,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return fmt.Errorf("encode link token request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/telegram/link-token", &buf)
	if err != nil {
		return fmt.Errorf("create link token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.internalKey == "" {
		return ErrUnauthorized
	}
	req.Header.Set("X-Internal-Key", c.internalKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send link token request: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		body := readBodySnippet(resp.Body)
		return fmt.Errorf("%w: status=%d body=%s", ErrBadRequest, resp.StatusCode, body)
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		body := readBodySnippet(resp.Body)
		return fmt.Errorf("%w: status=%d body=%s", ErrDeliveryFailed, resp.StatusCode, body)
	}
}

func (c *HTTPClient) SendOTP(ctx context.Context, phone, otpCode string) error {
	if phone == "" {
		return fmt.Errorf("%w: phone is required", ErrDeliveryFailed)
	}
	if otpCode == "" {
		return fmt.Errorf("%w: otp code is required", ErrDeliveryFailed)
	}
	payload := struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}{
		Phone: phone,
		Code:  otpCode,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return fmt.Errorf("encode otp request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/otp/send", &buf)
	if err != nil {
		return fmt.Errorf("create otp request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.internalKey == "" {
		return ErrUnauthorized
	}
	req.Header.Set("X-Internal-Key", c.internalKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send otp request: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return ErrNotLinked
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		body := readBodySnippet(resp.Body)
		return fmt.Errorf("%w: status=%d body=%s", ErrDeliveryFailed, resp.StatusCode, body)
	}
}

func readBodySnippet(r io.Reader) string {
	body, err := io.ReadAll(io.LimitReader(r, 4096))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}
