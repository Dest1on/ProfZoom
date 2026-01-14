package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Service отправляет сообщения в чаты Telegram.
type Service interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

// MarkupSender отправляет сообщения с reply markup.
type MarkupSender interface {
	SendMessageWithMarkup(ctx context.Context, chatID int64, text string, replyMarkup any) error
}

// Client реализует Service через Telegram Bot API.
type Client struct {
	botToken   string
	baseURL    string
	httpClient *http.Client
}

// APIError описывает ответ не 2xx от Telegram API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("telegram send: unexpected status %d", e.StatusCode)
	}
	return fmt.Sprintf("telegram send: unexpected status %d: %s", e.StatusCode, e.Body)
}

// NewClient создает Telegram-клиент с усиленным HTTP-клиентом.
func NewClient(botToken string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Client{
		botToken:   botToken,
		baseURL:    "https://api.telegram.org",
		httpClient: httpClient,
	}
}

// SendMessage отправляет текстовое сообщение в чат.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	return c.SendMessageWithMarkup(ctx, chatID, text, nil)
}

// SendMessageWithMarkup отправляет сообщение в чат с опциональным reply markup.
func (c *Client) SendMessageWithMarkup(ctx context.Context, chatID int64, text string, replyMarkup any) error {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram marshal payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return &APIError{StatusCode: resp.StatusCode, Body: string(payload)}
	}

	return nil
}
